/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package entities

import (
	"encoding/json"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"strconv"
)

// copySecurityRule returns a copy of a given securityRule
func copySecurityRule (rule *grpc_application_go.SecurityRule) * grpc_application_go.SecurityRule {
	if rule == nil {
		return nil
	}
	return &grpc_application_go.SecurityRule{
		OrganizationId: rule.OrganizationId,
		AppDescriptorId: rule.AppDescriptorId,
		RuleId: rule.RuleId,
		Name: rule.Name,
		TargetServiceName: rule.TargetServiceName,
		TargetPort: rule.TargetPort,
		Access: rule.Access,
		AuthServiceGroupName: rule.AuthServiceGroupName,
		AuthServices :rule.AuthServices,
		DeviceGroupIds : rule.DeviceGroupIds,
		DeviceGroupNames: rule.DeviceGroupNames,
	}
}

// copyServiceGroup returns a copy of a given ServiceGroup
func copyServiceGroup(group *grpc_application_go.ServiceGroup) *grpc_application_go.ServiceGroup {

	if group == nil {
		return nil
	}

	services := make ([]*grpc_application_go.Service, 0)
	for _, service := range group.Services {
		services = append(services, copyService(service))
	}

	return &grpc_application_go.ServiceGroup{
		OrganizationId: group.OrganizationId,
		AppDescriptorId: group.AppDescriptorId,
		ServiceGroupId: group.ServiceGroupId,
		Name: group.Name,
		Services: services,
		Policy: group.Policy,
		//Specs *ServiceGroupDeploymentSpecs `protobuf:"bytes,8,opt,name=specs,proto3" json:"specs,omitempty"`
		Labels: group.Labels,
	}
}

// copyService returns a copy of a given Service
func copyService (service * grpc_application_go.Service) * grpc_application_go.Service{

	if service == nil {
		return nil
	}

	storage := make ([]*grpc_application_go.Storage, 0)
	for _, sto := range service.Storage {
		storage = append(storage, copyStorage(sto))
	}

	ports := make ([]*grpc_application_go.Port, 0)
	for _, port := range service.ExposedPorts {
		ports = append(ports, copyExposedPort(port))
	}

	files := make ([]*grpc_application_go.ConfigFile, 0)
	for _, file := range service.Configs {
		files = append(files, copyConfigFile(file))
	}

	return &grpc_application_go.Service {
		OrganizationId: service.OrganizationId,
		AppDescriptorId: service.AppDescriptorId,
		ServiceGroupId: service.ServiceGroupId,
		ServiceId: service.ServiceId,
		Name: service.Name,
		Type: service.Type,
		Image: service.Image,
		Credentials: copyImageCredential(service.Credentials),
		Specs: copyDeploySpec(service.Specs),
		Storage: storage,
		ExposedPorts: ports,
		EnvironmentVariables: service.EnvironmentVariables,
		Configs: files,
		Labels: service.Labels,
		DeployAfter: service.DeployAfter,
		RunArguments: service.RunArguments,
	}

}
// copyStorage returns a copy of a given Storage
func copyStorage (storage * grpc_application_go.Storage) * grpc_application_go.Storage {
	if storage == nil {
		return nil
	}
	return &grpc_application_go.Storage{
		Size: storage.Size,
		MountPath: storage.MountPath,
		Type: storage.Type,
	}
}

// copyImageCredentials returns a copy of a given Image Credentiall
func copyImageCredential (credentials * grpc_application_go.ImageCredentials) * grpc_application_go.ImageCredentials {
	if credentials == nil {
		return nil
	}
	return &grpc_application_go.ImageCredentials{
		Username: 	credentials.Username,
		Password: 	credentials.Password,
		Email: 		credentials.Email,
		DockerRepository: credentials.DockerRepository,
	}
}

// copyDeploySpecs returns a copy of a given deploy spec
func copyDeploySpec (specs * grpc_application_go.DeploySpecs) * grpc_application_go.DeploySpecs {
	if specs == nil {
		return nil
	}
	return &grpc_application_go.DeploySpecs{
		Cpu: specs.Cpu,
		Memory: specs.Memory,
		Replicas: specs.Replicas,
	}
}

// copyExposedPort returns a copy of a given exposed port
func copyExposedPort (port * grpc_application_go.Port) * grpc_application_go.Port {
	if port == nil {
		return nil
	}

	endpoints := make([]*grpc_application_go.Endpoint, 0)
	for _, endpoint := range port.Endpoints{
		endpoints = append(endpoints, copyEndPoint(endpoint))
	}

	return &grpc_application_go.Port {
		Name: port.Name,
		InternalPort: port.InternalPort,
		ExposedPort: port.ExposedPort,
		Endpoints: endpoints,
	}
}

// copyEndPoint returns a copy of a given endpoint
func copyEndPoint (endpoint *grpc_application_go.Endpoint) * grpc_application_go.Endpoint {
 	if endpoint == nil {
 		return nil
	}
 	return &grpc_application_go.Endpoint{
 		Type: endpoint.Type,
 		Path: endpoint.Path,
	}
}

// copyConfigFile returns a copy of a given configFile
func copyConfigFile (configFile * grpc_application_go.ConfigFile) * grpc_application_go.ConfigFile {
	if configFile == nil {
		return nil
	}
	return &grpc_application_go.ConfigFile{
		OrganizationId: configFile.OrganizationId,
		AppDescriptorId: configFile.AppDescriptorId,
		ConfigFileId: configFile.ConfigFileId,
		Name: configFile.Name,
		Content: configFile.Content,
		MountPath: configFile.MountPath,
	}
}

// newParametrizedDescriptorFromDescriptor returns a parameterized descriptor as a copy of a given descriptor
func newParametrizedDescriptorFromDescriptor(descriptor *grpc_application_go.AppDescriptor) *grpc_application_manager_go.ParametrizedDescriptor {
	if descriptor == nil {
		return nil
	}
	rules := make ([]*grpc_application_go.SecurityRule, 0)
	for _, rule := range descriptor.Rules {
		rules = append(rules, copySecurityRule(rule))
	}
	groups := make ([]*grpc_application_go.ServiceGroup, 0)
	for _, group := range descriptor.Groups {
		groups = append(groups, copyServiceGroup(group))
	}


	return &grpc_application_go.ParametrizedDescriptor{
		OrganizationId: descriptor.OrganizationId,
		AppDescriptorId: descriptor.AppDescriptorId,
		Name: descriptor.Name,
		ConfigurationOptions: descriptor.ConfigurationOptions,
		EnvironmentVariables: descriptor.EnvironmentVariables,
		Labels: descriptor.Labels,
		Rules: rules,
		Groups: groups,
	}
}

// findParameterInDescriptor looks for the definition of a given instance parameter in the description of the descriptor
func findParameterInDescriptor(descriptor *grpc_application_go.AppDescriptor,
	parameter grpc_application_go.InstanceParameter) (*grpc_application_go.AppParameter, derrors.Error){

	if descriptor.Parameters == nil || len(descriptor.Parameters) == 0 {
		return nil, derrors.NewInvalidArgumentError("Instance parameter not found. Descriptor has no parameters")
	}

	for _, param := range descriptor.Parameters {
		if param.Name == parameter.ParameterName {
			return param, nil
		}
	}

	return nil, derrors.NewNotFoundError("Instance parameter not found in descriptor definition").WithParams(parameter.ParameterName)
}

// applyParameter substitutes the entry of the descriptor for the indicated value
func applyParameter (jsonParamDescriptor *string,
	paramDefinition grpc_application_go.AppParameter,
	value interface{}) (derrors.Error){

	path := paramDefinition.Path

	old := gjson.Get(*jsonParamDescriptor, path)
	if old.Raw != "" {

	}

	// https://github.com/tidwall/sjson
	json, err := sjson.Set(*jsonParamDescriptor, path, value)
	if err != nil {
		return  conversions.ToDerror(err)
	}

	*jsonParamDescriptor = json
	return  nil
}

// validateInstanceParameter validates that the type of the value parameter matches that of the description of the parameter in the descriptor
func validateInstanceParameter (paramDefinition grpc_application_go.AppParameter,
	parameter grpc_application_go.InstanceParameter) (interface{}, derrors.Error) {

		var value interface{}
		var err error
		// validate type
		switch paramDefinition.Type {
		case grpc_application_go.ParamDataType_BOOLEAN:
			value, err = strconv.ParseBool(parameter.Value)
			if err != nil {
				return nil, conversions.ToDerror(err)
			}
		case grpc_application_go.ParamDataType_INTEGER:
			value, err = strconv.Atoi(parameter.Value)
			if err != nil {
				return nil, conversions.ToDerror(err)
			}
		case grpc_application_go.ParamDataType_FLOAT:
			value, err = strconv.ParseFloat(parameter.Value, 32)
			if err != nil {
				return nil, conversions.ToDerror(err)
			}
		case grpc_application_go.ParamDataType_ENUM:
			find := false
			for _, paramVal := range paramDefinition.EnumValues {
				if paramVal == parameter.Value {
					find = true
					break
				}
			}
			if ! find {
				return nil, derrors.NewInvalidArgumentError("Invalid parameter value").WithParams("parameter", parameter.ParameterName).WithParams("value", parameter.Value)
			}
			value = parameter.Value
		case grpc_application_go.ParamDataType_STRING:
			value = parameter.Value
		case grpc_application_go.ParamDataType_PASSWORD:
			value = parameter.Value

		}


		return value, nil
}

// CreateParametrizedDescriptor returns a parameterized descriptor once the parameters of the instance
// have been validated and applied to the given descriptor
func CreateParametrizedDescriptor (descriptor *grpc_application_go.AppDescriptor,
	parameters *grpc_application_go.InstanceParameterList) (*grpc_application_go.ParametrizedDescriptor, derrors.Error) {

		parametrized := newParametrizedDescriptorFromDescriptor(descriptor)

		if parameters == nil || parameters.Parameters == nil || len(parameters.Parameters) == 0 {
			return parametrized, nil
		}

		// we need to convert the parametrized descriptor to json to apply changes
		newDescriptor, err:= json.Marshal(parametrized)
		if err != nil {
			return nil, conversions.ToDerror(err)
		}

		jsonDescriptor := string(newDescriptor)

		for _, param := range parameters.Parameters {

			// find parameter definition, if the parameter does no exists an error is returned
			paramDefinition, err := findParameterInDescriptor(descriptor, *param)
			if err != nil {
				return nil, err
			}

			// validate parameter
			value, err := validateInstanceParameter(*paramDefinition, *param)
			if err != nil {
				return nil, err
			}
			// apply
			err = applyParameter(&jsonDescriptor, *paramDefinition, value)
			if err != nil {
				return nil, err
			}
		}

		// convert json to parametrizedDescriptor
	 	err = json.Unmarshal ([]byte(jsonDescriptor), parametrized)
	 	if err != nil {
	 		return nil, conversions.ToDerror(err)
		}


		return parametrized, nil
}
