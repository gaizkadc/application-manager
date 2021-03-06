/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package application

import (
	"context"
	"fmt"
	"github.com/nalej/application-manager/internal/pkg/entities"
	appnet "github.com/nalej/application-manager/internal/pkg/server/application-network"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-application-network-go"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-device-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-organization-manager-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/nalej-bus/pkg/queue/application/ops"
	"github.com/rs/zerolog/log"
	"math/rand"
	"sync"
	"time"
)

const DefaultTimeout = time.Minute
const RequiredParamNotFilled = "Required parameter not filled"
const RequiredOutboundNotFilled = "Required outbound not filled"
const OutboundNotDefined = "Deploy outbound connection not defined"

// Manager structure with the required clients for roles operations.
type Manager struct {
	appClient       grpc_application_go.ApplicationsClient
	orgClient       grpc_organization_manager_go.OrganizationsClient
	conductorClient grpc_conductor_go.ConductorClient
	clusterClient   grpc_infrastructure_go.ClustersClient
	deviceClient    grpc_device_go.DevicesClient
	appNetClient    grpc_application_network_go.ApplicationNetworkClient
	appOpsProducer  *ops.ApplicationOpsProducer
	appNetManager   appnet.Manager
}

// NewManager creates a Manager using a set of clients.
func NewManager(
	appClient grpc_application_go.ApplicationsClient,
	orgClient grpc_organization_manager_go.OrganizationsClient,
	conductorClient grpc_conductor_go.ConductorClient,
	clusterClient grpc_infrastructure_go.ClustersClient,
	deviceClient grpc_device_go.DevicesClient,
	appNetClient grpc_application_network_go.ApplicationNetworkClient,
	appOpsProducer *ops.ApplicationOpsProducer,
	appNetManager appnet.Manager) Manager {
	return Manager{appClient, orgClient, conductorClient, clusterClient, deviceClient, appNetClient, appOpsProducer, appNetManager}
}

// AddAppDescriptor adds a new application descriptor to a given organization.
func (m *Manager) AddAppDescriptor(addDescriptorRequest *grpc_application_go.AddAppDescriptorRequest) (*grpc_application_go.AppDescriptor, error) {

	// before add appDescriptor, validate parameters
	err := entities.ValidateDescriptorParameters(addDescriptorRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}

	return m.appClient.AddAppDescriptor(context.Background(), addDescriptorRequest)
}

// ListAppDescriptors retrieves a list of application descriptors.
func (m *Manager) ListAppDescriptors(organizationID *grpc_organization_go.OrganizationId) (*grpc_application_go.AppDescriptorList, error) {
	return m.appClient.ListAppDescriptors(context.Background(), organizationID)
}

// GetAppDescriptor retrieves a given application descriptor.
func (m *Manager) GetAppDescriptor(appDescriptorID *grpc_application_go.AppDescriptorId) (*grpc_application_go.AppDescriptor, error) {
	return m.appClient.GetAppDescriptor(context.Background(), appDescriptorID)
}

// UpdateAppDescriptor allows the user to update the information of a registered descriptor.
func (m *Manager) UpdateAppDescriptor(request *grpc_application_go.UpdateAppDescriptorRequest) (*grpc_application_go.AppDescriptor, error) {
	return m.appClient.UpdateAppDescriptor(context.Background(), request)
}

// RemoveAppDescriptor removes an application descriptor from the system.
func (m *Manager) RemoveAppDescriptor(appDescriptorID *grpc_application_go.AppDescriptorId) (*grpc_common_go.Success, error) {
	// Check if there are instances running with that descriptor
	orgID := &grpc_organization_go.OrganizationId{
		OrganizationId: appDescriptorID.OrganizationId,
	}
	instances, err := m.appClient.ListAppInstances(context.Background(), orgID)
	if err != nil {
		return nil, err
	}
	for _, inst := range instances.Instances {
		if inst.AppDescriptorId == appDescriptorID.AppDescriptorId {
			return nil, derrors.NewFailedPreconditionError("application instances must be removed before deleting the descriptor")
		}
	}
	return m.appClient.RemoveAppDescriptor(context.Background(), appDescriptorID)
}

// checkAllRequiredParametersAreFilled checks all the params defined as required are filled in deploy request
func (m *Manager) checkAllRequiredParametersAreFilled(desc *grpc_application_go.AppDescriptor, params *grpc_application_go.InstanceParameterList) error {
	// get all the required parameters
	for _, p := range desc.Parameters {
		if p.Required == true {
			find := false
			if params != nil {
				// look for it in deploy params
				for _, deployParam := range params.Parameters {

					if deployParam.ParameterName == p.Name {
						find = true
						break
					}

				}
			}
			if !find {
				return derrors.NewFailedPreconditionError(RequiredParamNotFilled)
			}
		}
	}

	return nil
}

// CheckInboundResponse struct that contains the result of checkInbound operation.
type CheckInboundResponse struct {
	// InstanceId with the targetInstance identifier
	InstanceId string
	// InboundName with the name of the inbound not found
	InboundName string
	// Result contains the result of the validation
	Result bool
}

// checkInbounds checks if the instanceID has defined all the inbounds in the inboundNames array
func (m *Manager) checkInbounds(respond chan<- CheckInboundResponse, wg *sync.WaitGroup, organizationID string, instanceID string, inboundNames []string) {
	defer wg.Done()

	log.Debug().Str("TargetInstanceId", instanceID).Interface("TargetInboundNames", inboundNames).Msg("check inbounds Interface")

	targetInstance, err := m.appClient.GetAppInstance(context.Background(),
		&grpc_application_go.AppInstanceId{
			OrganizationId: organizationID,
			AppInstanceId:  instanceID,
		})
	if err != nil {
		respond <- CheckInboundResponse{
			InstanceId: instanceID,
			Result:     false,
		}
		return
	}
	// globalResult contains the result of the operation,
	// it is true when ALL the names are found
	// it is false when one of them is not found
	globalResult := true
	for _, inboundName := range inboundNames {
		targetFound := false
		for _, inbound := range targetInstance.InboundNetInterfaces {
			if inbound.Name == inboundName {
				targetFound = true
			}
		}
		if !targetFound {
			respond <- CheckInboundResponse{
				InstanceId:  instanceID,
				InboundName: inboundName,
				Result:      false,
			}
			globalResult = false
			break
		}
	}
	// if all is OK -> true result sent to the chan
	if globalResult {
		respond <- CheckInboundResponse{
			InstanceId: instanceID,
			Result:     true,
		}
	}

}

// checkConnections: Checks all the connection fields are consistent (the target_instance_id has an inbound named TargetInboundName)
// and checks the required outbounds are informed
func (m *Manager) checkConnections(organizationID string, connections []*grpc_application_manager_go.ConnectionRequest,
	outboundInterfaces []*grpc_application_go.OutboundNetworkInterface) derrors.Error {

	// 1.- Check required outbounds
	for _, outbound := range outboundInterfaces {
		if outbound.Required {
			log.Debug().Interface("outboundName", outbound.Name).Msg("check required outbound")
			found := false
			for _, connection := range connections {
				if connection.SourceOutboundName == outbound.Name {
					found = true
				}
			}
			if !found {
				return derrors.NewFailedPreconditionError(RequiredOutboundNotFilled).WithParams(outbound.Name)
			}
		}
	}

	// create a map with all the inbounds per instance_id
	instanceList := make(map[string][]string, 0)
	for _, conn := range connections {
		inbounds, exists := instanceList[conn.TargetInstanceId]
		if !exists {
			instanceList[conn.TargetInstanceId] = []string{conn.TargetInboundName}
		} else {
			instanceList[conn.TargetInstanceId] = append(inbounds, conn.TargetInboundName)
		}
	}

	respond := make(chan CheckInboundResponse, len(instanceList))
	var wg sync.WaitGroup

	wg.Add(len(instanceList))

	for instanceId, inboundList := range instanceList {
		log.Debug().Str("instanceID", instanceId).Interface("inboundList", inboundList).Msg("check inbound names")

		go m.checkInbounds(respond, &wg, organizationID, instanceId, inboundList)

	}

	wg.Wait()
	close(respond)

	for result := range respond {
		if result.Result == false {
			if result.InboundName != "" {
				return derrors.NewFailedPreconditionError("no inbound interface found").WithParams(result.InstanceId, result.InboundName)
			} else { // database error getting the instanceId (or instanceID does not exist)
				return derrors.NewFailedPreconditionError("instance not found").WithParams(result.InstanceId)
			}
		}
	}

	return nil

}

// Deploy an application descriptor.
func (m *Manager) Deploy(deployRequest *grpc_application_manager_go.DeployRequest) (*grpc_application_manager_go.DeploymentResponse, error) {

	log.Debug().Interface("request", deployRequest).Msg("received deployment request")

	// Retrieve descriptor by descriptorID
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	desc, err := m.appClient.GetAppDescriptor(ctx, &grpc_application_go.AppDescriptorId{
		OrganizationId:  deployRequest.OrganizationId,
		AppDescriptorId: deployRequest.AppDescriptorId,
	})
	if err != nil {
		log.Error().Err(err).Msgf("error getting application descriptor %s", deployRequest.AppDescriptorId)
		return nil, err
	}

	// check if all required params are filled
	err = m.checkAllRequiredParametersAreFilled(desc, deployRequest.Parameters)
	if err != nil {
		return nil, err
	}

	// NP-1963. Check connections
	// 1.- TargetInstanceId has an inbound named TargetInboundName
	// 2.- The descriptor has an outbound named SourceOutboundName
	// 3.- All required outbound are informed
	dErr := m.checkConnections(deployRequest.OrganizationId, deployRequest.OutboundConnections, desc.OutboundNetInterfaces)
	if dErr != nil {
		return nil, conversions.ToGRPCError(dErr)
	}

	orgSettings := entities.NewOrganizationSettings(deployRequest.OrganizationId, m.orgClient)

	// Create it parametrized descriptor
	parametrizedDesc, err := entities.CreateParametrizedDescriptor(desc, deployRequest.Parameters, orgSettings)
	if err != nil {
		log.Error().Err(err).Msgf("error creating  parametrized descriptor %s.", deployRequest.AppDescriptorId)
		return nil, err
	}

	// Create new application instance
	addReq := &grpc_application_go.AddAppInstanceRequest{
		OrganizationId:  deployRequest.OrganizationId,
		AppDescriptorId: deployRequest.AppDescriptorId,
		Name:            deployRequest.Name,
		Parameters:      deployRequest.Parameters,
	}

	// Add instance, by default this is created with bus status
	ctxInstance, cancelInstance := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelInstance()
	instance, err := m.appClient.AddAppInstance(ctxInstance, addReq)
	if err != nil {
		log.Error().Err(err).Msg("error adding application instance")
		return nil, err
	}

	connections := make([]*grpc_application_network_go.ConnectionInstance, len(deployRequest.OutboundConnections))
	for connectionIndex, connectionRequest := range deployRequest.OutboundConnections {
		sourceInstanceName := ""
		// TODO Too cumbersome. Consider a refactor of the descriptor to link outbound interfaces to services explicitly
		for _, rule := range instance.Rules {
			if rule.OutboundNetInterface == connectionRequest.SourceOutboundName {
				sourceInstanceName = rule.TargetServiceName
			}
		}
		if sourceInstanceName == "" {
			log.Error().Interface("connectionRequest", connectionRequest).Msg("the connection request refers to an outbound interface name not linked to a service. Skipping.")
			continue
		}
		connections[connectionIndex] = &grpc_application_network_go.ConnectionInstance{
			OrganizationId:     desc.OrganizationId,
			SourceInstanceName: sourceInstanceName,
			TargetInstanceId:   connectionRequest.TargetInstanceId,
			InboundName:        connectionRequest.TargetInboundName,
			OutboundName:       connectionRequest.SourceOutboundName,
		}
	}

	// fill the instance_id in the parametrized descriptor
	parametrizedDesc.AppInstanceId = instance.AppInstanceId

	appInstanceID := &grpc_application_go.AppInstanceId{
		OrganizationId: deployRequest.OrganizationId,
		AppInstanceId:  instance.AppInstanceId,
	}

	// Add parametrizedDescriptor in the system
	ctxParametrized, cancelParametrized := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelParametrized()
	newDesc, err := m.appClient.AddParametrizedDescriptor(ctxParametrized, parametrizedDesc)
	if err != nil {
		log.Error().Err(err).Msgf("error adding  parametrized descriptor %s. Delete instance", instance.AppInstanceId)
		_, rollbackErr := m.appClient.RemoveAppInstance(context.Background(), appInstanceID)
		if rollbackErr != nil {
			log.Error().Err(err).Msgf("error in rollback deleting the instance %s", instance.AppInstanceId)
		}
		return nil, err
	}

	// update the instance with the rules parametrized
	if len(parametrizedDesc.Rules) > 0 {
		ctxUpdateInstance, cancelUpdate := context.WithTimeout(context.Background(), DefaultTimeout)
		defer cancelUpdate()
		// update the instance
		instance.Rules = newDesc.Rules
		instance.ConfigurationOptions = newDesc.ConfigurationOptions
		instance.EnvironmentVariables = newDesc.EnvironmentVariables
		instance.Labels = newDesc.Labels
		_, err := m.appClient.UpdateAppInstance(ctxUpdateInstance, instance)

		if err != nil {
			log.Error().Err(err).Msgf("error updating instance %s. Delete instance", instance.AppInstanceId)
			_, rollbackErr := m.appClient.RemoveAppInstance(context.Background(), appInstanceID)
			if rollbackErr != nil {
				log.Error().Err(err).Msgf("error in rollback deleting the instance %s", instance.AppInstanceId)
			}
			return nil, err
		}

	}

	// send deploy command to conductor
	request := &grpc_conductor_go.DeploymentRequest{
		RequestId:           fmt.Sprintf("app-mngr-%d", rand.Int()),
		AppInstanceId:       appInstanceID,
		Name:                deployRequest.Name,
		OutboundConnections: connections,
	}

	ctx, cancel = context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	err = m.appOpsProducer.Send(ctx, request)
	if err != nil {
		log.Error().Err(err).Str("appInstanceId", instance.AppInstanceId).
			Msg("error when sending deployment request to the queue")
		return nil, err
	}

	toReturn := grpc_application_manager_go.DeploymentResponse{
		RequestId:     fmt.Sprintf("app-mngr-%d", rand.Int()),
		AppInstanceId: instance.AppInstanceId,
		Status:        grpc_application_go.ApplicationStatus_QUEUED}

	log.Debug().Interface("deploymentResponse", toReturn).Msg("Response")

	return &toReturn, nil

}

// Undeploy a running application instance.
func (m *Manager) Undeploy(undeployRequest *grpc_application_manager_go.UndeployRequest) (*grpc_common_go.Success, error) {

	// GetAppInstance returns expanded instance (with its connections)
	instance, iErr := m.GetAppInstance(&grpc_application_go.AppInstanceId{
		OrganizationId: undeployRequest.OrganizationId,
		AppInstanceId:  undeployRequest.AppInstanceId,
	})

	if iErr != nil {
		log.Error().Err(iErr).Str("appInstanceId", undeployRequest.AppInstanceId).
			Msg("error when sending the undeploy request to the queue")
		return nil, iErr
	}

	if len(instance.InboundConnections) > 0 && !undeployRequest.UserConfirmation {
		return nil, conversions.ToGRPCError(derrors.NewFailedPreconditionError("can not undeploy the instance, it has inbound connections. User confirmation required"))
	}

	// Remove Inbound connections
	for _, conn := range instance.InboundConnections {
		_, rErr := m.appNetManager.RemoveConnection(&grpc_application_network_go.RemoveConnectionRequest{
			OrganizationId:   conn.OrganizationId,
			SourceInstanceId: conn.SourceInstanceId,
			TargetInstanceId: conn.TargetInstanceId,
			InboundName:      conn.InboundName,
			OutboundName:     conn.OutboundName,
			UserConfirmation: true,
		})
		if rErr != nil {
			// I think we can continue undeploying the instance, the ztNetwork is going to be deleted and the namespaces too.
			// Some "invalid" records will remain in the database
			log.Error().Err(rErr).Msg("Error removing inbound connection")
		}
	}
	for _, conn := range instance.OutboundConnections {
		_, rErr := m.appNetManager.RemoveConnection(&grpc_application_network_go.RemoveConnectionRequest{
			OrganizationId:   conn.OrganizationId,
			SourceInstanceId: conn.SourceInstanceId,
			TargetInstanceId: conn.TargetInstanceId,
			InboundName:      conn.InboundName,
			OutboundName:     conn.OutboundName,
			UserConfirmation: true,
		})
		if rErr != nil {
			// I think we can continue undeploying the instance, the ztNetwork is going to be deleted and the namespaces too.
			// Some "invalid" records will remain in the database
			log.Error().Err(rErr).Msg("Error removing outbound connection")
		}
	}

	appInstanceID := &grpc_conductor_go.UndeployRequest{
		OrganizationId: undeployRequest.OrganizationId,
		AppInstanceId:  undeployRequest.AppInstanceId,
	}
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	err := m.appOpsProducer.Send(ctx, appInstanceID)
	if err != nil {
		log.Error().Err(err).Str("appInstanceId", undeployRequest.AppInstanceId).
			Msg("error when sending the undeploy request to the queue")
		return nil, err
	}

	return &grpc_common_go.Success{}, nil

}

// getInstanceConnections returns the appInstance with the connections field filled
func (m *Manager) getInstanceConnections(instance *grpc_application_go.AppInstance) *grpc_application_manager_go.AppInstance {

	expandInstance := entities.ToAppInstance(instance)

	appInstanceID := &grpc_application_go.AppInstanceId{
		OrganizationId: instance.OrganizationId,
		AppInstanceId:  instance.AppInstanceId,
	}

	// InboundConnections
	inboundConnections, err := m.appNetClient.ListInboundConnections(context.Background(), appInstanceID)
	if err != nil {
		log.Error().Str("instance_id", instance.AppInstanceId).Msg("error getting inbound connections")
	} else {
		if inboundConnections != nil {
			expandInstance.InboundConnections = inboundConnections.Connections
		}
	}

	// OutboundConnections
	outboundConnections, err := m.appNetClient.ListOutboundConnections(context.Background(), appInstanceID)
	if err != nil {
		log.Error().Str("instance_id", instance.AppInstanceId).Msg("error getting outbound connections")
	} else {
		if outboundConnections != nil {
			expandInstance.OutboundConnections = outboundConnections.Connections
		}
	}

	return expandInstance

}

// ListAppInstances retrieves a list of application descriptors.
func (m *Manager) ListAppInstances(organizationID *grpc_organization_go.OrganizationId) (*grpc_application_manager_go.AppInstanceList, error) {

	list, err := m.appClient.ListAppInstances(context.Background(), organizationID)
	if err != nil {
		return nil, err
	}
	expandList := make([]*grpc_application_manager_go.AppInstance, 0)
	for _, instance := range list.Instances {
		expandList = append(expandList, m.getInstanceConnections(instance))
	}
	return &grpc_application_manager_go.AppInstanceList{
		Instances: expandList,
	}, nil
}

// GetAppDescriptor retrieves a given application descriptor.
func (m *Manager) GetAppInstance(appInstanceID *grpc_application_go.AppInstanceId) (*grpc_application_manager_go.AppInstance, error) {

	appInstance, err := m.appClient.GetAppInstance(context.Background(), appInstanceID)

	if err != nil {
		return nil, err
	}

	// get inbound and outbound connections for the instance
	expandInstance := m.getInstanceConnections(appInstance)
	return expandInstance, nil
}

func (m *Manager) ListInstanceParameters(appInstanceID *grpc_application_go.AppInstanceId) (*grpc_application_go.InstanceParameterList, error) {
	return m.appClient.GetInstanceParameters(context.Background(), appInstanceID)
}

func (m *Manager) ListDescriptorAppParameters(descriptorID *grpc_application_go.AppDescriptorId) (*grpc_application_go.AppParameterList, error) {
	return m.appClient.GetDescriptorAppParameters(context.Background(), descriptorID)
}

func (m *Manager) RetrieveTargetApplications(filter *grpc_application_manager_go.ApplicationFilter) (*grpc_application_manager_go.TargetApplicationList, error) {

	// check if the device_group_id and device_group_name are correct
	group, err := m.deviceClient.GetDeviceGroup(context.Background(), &grpc_device_go.DeviceGroupId{
		OrganizationId: filter.OrganizationId,
		DeviceGroupId:  filter.DeviceGroupId,
	})
	if err != nil {
		return nil, err
	}
	if group.Name != filter.DeviceGroupName {
		return nil, conversions.ToGRPCError(derrors.NewPermissionDeniedError("cannot access device_group_name"))
	}

	orgID := &grpc_organization_go.OrganizationId{
		OrganizationId: filter.OrganizationId,
	}
	// TODO allow filtering on the list request
	allApps, err := m.appClient.ListAppInstances(context.Background(), orgID)
	if err != nil {
		return nil, err
	}

	filtered := ApplyFilter(allApps, filter)

	result, fErr := ToApplicationLabelsList(filtered)
	if fErr != nil {
		return nil, conversions.ToGRPCError(fErr)
	}
	return result, nil
}

func (m *Manager) fillEndpoints(endpoints []*grpc_application_go.EndpointInstance) {
	for i := 0; i < len(endpoints); i++ {
		endpoints[i].Fqdn = fmt.Sprintf("%s:%d", endpoints[i].Fqdn, endpoints[i].Port)
	}
}

func (m *Manager) RetrieveEndpoints(request *grpc_application_manager_go.RetrieveEndpointsRequest) (*grpc_application_manager_go.ApplicationEndpoints, error) {

	instanceID := &grpc_application_go.AppInstanceId{
		OrganizationId: request.OrganizationId,
		AppInstanceId:  request.AppInstanceId,
	}
	// get the instance requested
	instance, err := m.appClient.GetAppInstance(context.Background(), instanceID)
	if err != nil {
		return nil, err
	}

	appClusterEndPoints := make([]*grpc_application_manager_go.ApplicationClusterEndpoints, 0)

	//foreach serviceInstance in appInstance -> get endPoints and DeployedClusterId
	for _, group := range instance.Groups {
		for _, service := range group.ServiceInstances {

			// get the clusterHost (if the service is RUNNING)
			if service.Status == grpc_application_go.ServiceStatus_SERVICE_RUNNING &&
				len(service.Endpoints) > 0 { // the service has endpoints

				clusterId := &grpc_infrastructure_go.ClusterId{
					OrganizationId: request.OrganizationId,
					ClusterId:      service.DeployedOnClusterId,
				}
				cluster, err := m.clusterClient.GetCluster(context.Background(), clusterId)
				if err != nil {
					return nil, err
				}

				m.fillEndpoints(service.Endpoints)

				clusterEndPoint := &grpc_application_manager_go.ApplicationClusterEndpoints{
					DeviceControllerUrl: fmt.Sprintf("device-controller.%s", cluster.Hostname),
					Endpoints:           service.Endpoints,
				}
				appClusterEndPoints = append(appClusterEndPoints, clusterEndPoint)
			}
		}
	}

	return &grpc_application_manager_go.ApplicationEndpoints{
		ClusterEndpoints: appClusterEndPoints,
	}, nil

}

// ListAvailableInstanceInbounds List all the pluggable inbounds
func (m *Manager) ListAvailableInstanceInbounds(organizationId *grpc_organization_go.OrganizationId) (*grpc_application_manager_go.AvailableInstanceInboundList, error) {
	appInstances, err := m.appClient.ListAppInstances(context.Background(), organizationId)
	if err != nil {
		return nil, err
	}
	instanceInbounds := make([]*grpc_application_manager_go.AvailableInstanceInbound, 0)
	for _, appInstance := range appInstances.Instances {
		for _, inbound := range appInstance.InboundNetInterfaces {
			instanceInbounds = append(instanceInbounds, &grpc_application_manager_go.AvailableInstanceInbound{
				OrganizationId: organizationId.OrganizationId,
				AppInstanceId:  appInstance.AppInstanceId,
				InstanceName:   appInstance.Name,
				InboundName:    inbound.Name,
			})
		}
	}
	return &grpc_application_manager_go.AvailableInstanceInboundList{InstanceInbounds: instanceInbounds}, nil
}

// ListAvailableInstanceOutbounds List all the outbounds that are not connected
func (m *Manager) ListAvailableInstanceOutbounds(organizationId *grpc_organization_go.OrganizationId) (*grpc_application_manager_go.AvailableInstanceOutboundList, error) {
	appInstances, err := m.appClient.ListAppInstances(context.Background(), organizationId)
	if err != nil {
		return nil, err
	}
	instanceOutbounds := make([]*grpc_application_manager_go.AvailableInstanceOutbound, 0)
	for _, appInstance := range appInstances.Instances {
		expandedAppInstance := m.getInstanceConnections(appInstance)
		for _, outbound := range appInstance.OutboundNetInterfaces {
			connected := false
			for _, connection := range expandedAppInstance.OutboundConnections {
				if outbound.Name == connection.OutboundName {
					connected = true
				}
			}
			// Exclude the connected outbounds
			if !connected {
				instanceOutbounds = append(instanceOutbounds, &grpc_application_manager_go.AvailableInstanceOutbound{
					OrganizationId: organizationId.OrganizationId,
					AppInstanceId:  appInstance.AppInstanceId,
					InstanceName:   appInstance.Name,
					OutboundName:   outbound.Name,
				})
			}
		}

	}
	return &grpc_application_manager_go.AvailableInstanceOutboundList{InstanceOutbounds: instanceOutbounds}, nil
}
