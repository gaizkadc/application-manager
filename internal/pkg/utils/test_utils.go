/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package utils

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/nalej/grpc-application-go"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"google.golang.org/grpc"
	"os"
)

// RunIntegrationTests checks whether integration tests should be executed.
func RunIntegrationTests() bool {
	var runIntegration = os.Getenv("RUN_INTEGRATION_TEST")
	return runIntegration == "true"
}

func GetConnection(address string) *grpc.ClientConn {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	gomega.Expect(err).To(gomega.Succeed())
	return conn
}

func CreateTestAppInstance(organizationID string, appDescriptorID string, appInstanceID string, labels map[string]string, groups []string) *grpc_application_go.AppInstance {
	service := &grpc_application_go.ServiceInstance{
		OrganizationId:      "",
		AppDescriptorId:     "",
		AppInstanceId:       "",
		ServiceId:           "service1",
		Endpoints:           nil,
		DeployedOnClusterId: "",
	}
	sr := &grpc_application_go.SecurityRule{
		OrganizationId:       organizationID,
		AppDescriptorId:      appDescriptorID,
		RuleId:               "rule1",
		Name:                 "",
		TargetServiceGroupName: "g1",
		TargetServiceName: "service1",
		TargetPort:           80,
		Access:               grpc_application_go.PortAccess_DEVICE_GROUP,
		DeviceGroupNames:       groups,
		DeviceGroupIds:			groups,
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}

	groupInstance := &grpc_application_go.ServiceGroupInstance{
		OrganizationId:         organizationID,
		AppDescriptorId:        appDescriptorID,
		AppInstanceId:          appInstanceID,
		ServiceGroupId:         "g1",
		ServiceGroupInstanceId: "gi1",
		Name:                   "",
		ServiceInstances:       []*grpc_application_go.ServiceInstance{service},
		Policy:                 0,
		Status:                 0,
		Metadata:               nil,
		Specs:                  nil,
		Labels:                 nil,
	}

	return &grpc_application_go.AppInstance{
		OrganizationId:  organizationID,
		AppDescriptorId: appDescriptorID,
		AppInstanceId:   appInstanceID,
		Labels:          labels,
		Rules:           []*grpc_application_go.SecurityRule{sr},
		Groups: []*grpc_application_go.ServiceGroupInstance{groupInstance},
	}
}

func CreateTestAppInstanceRequest (organizationID string, appDescriptorID string) *grpc_application_go.AddAppInstanceRequest {

	return &grpc_application_go.AddAppInstanceRequest{
		OrganizationId:  organizationID,
		AppDescriptorId: appDescriptorID,
		Name:   "test",
	}
}

func CreateAddAppDescriptorRequest(organizationID string, groups []string, labels map[string]string) * grpc_application_go.AddAppDescriptorRequest{
	service := &grpc_application_go.Service{
		OrganizationId:       organizationID,
		ServiceId:            uuid.New().String(),
		Name:                 "service-test",
		Type:                 grpc_application_go.ServiceType_DOCKER,
		Image:                "nginx:1.12",
		Specs:                &grpc_application_go.DeploySpecs{
			Replicas:             1,
		},
	}
	rules := make([]*grpc_application_go.SecurityRule, 0)
	rule := &grpc_application_go.SecurityRule {
		OrganizationId: organizationID,
		Name: "SecurityRule (it)",
		Access: grpc_application_go.PortAccess_DEVICE_GROUP,
		DeviceGroupNames: groups,
	}
	rules = append(rules, rule)

	group := &grpc_application_go.ServiceGroup{
		OrganizationId:       organizationID,
		Name:                 "g1",
		Services:             []*grpc_application_go.Service{service},
		Policy:               0,
		Specs:                nil,
		Labels:               nil,
	}

	toAdd := &grpc_application_go.AddAppDescriptorRequest{
		RequestId:            fmt.Sprintf("application-manager-it-%d", ginkgo.GinkgoRandomSeed()),
		OrganizationId:       organizationID,
		Name:                 fmt.Sprintf("app-manager-it-%d", ginkgo.GinkgoRandomSeed()),
		ConfigurationOptions: nil,
		EnvironmentVariables: nil,
		Labels:               labels,
		Rules:                rules,
		Groups:               []*grpc_application_go.ServiceGroup{group},
	}
	return toAdd
}

func CreateFullAppDescriptor()* grpc_application_go.AddAppDescriptorRequest{

	return &grpc_application_go.AddAppDescriptorRequest {
		RequestId: 		uuid.New().String(),
		OrganizationId:	uuid.New().String(),
		Name: 			"descriptor-test",
		Rules: 			[]*grpc_application_go.SecurityRule{
			{
				Name: "rule1",
				TargetServiceGroupName: "g1",
				TargetServiceName:"service1",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g2",
				AuthServices: []string{"service3"},

			},
			{
				Name: "rule2",
				TargetServiceGroupName: "g2",
				TargetServiceName:"service3",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g1",
				AuthServices: []string{"service1", "service2"},
			},
		},
		Groups: 		[]*grpc_application_go.ServiceGroup{
			{
				Name:"g1",
				Services:[]*grpc_application_go.Service{
					{Name: "service1",},
					{Name: "service2",},},
				Specs: &grpc_application_go.ServiceGroupDeploymentSpecs{
					Replicas: 3,
					MultiClusterReplica: false,
				},
			},
			{
				Name:"g2",
				Services:[]*grpc_application_go.Service{
					{Name: "service3",},},
			},
		},
		EnvironmentVariables:map[string]string{"var1": "NALEJ_SERV_SERVICE1:2000", "var2": "NALEJ_SERV_SERVICE2"},
	}
}

func CreateAppDescriptorWithoutGroups()* grpc_application_go.AddAppDescriptorRequest{

	return &grpc_application_go.AddAppDescriptorRequest {
		RequestId: 		uuid.New().String(),
		OrganizationId:	uuid.New().String(),
		Name: 			"descriptor-test",
		Rules: 			[]*grpc_application_go.SecurityRule{
			{
				Name: "rule1",
				TargetServiceGroupName: "g1",
				TargetServiceName:"service1",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g2",
				AuthServices: []string{"service3"},

			},
			{
				Name: "rule2",
				TargetServiceGroupName: "g2",
				TargetServiceName:"service3",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g1",
				AuthServices: []string{"service1", "service2"},
			},
		},
		Groups: 		[]*grpc_application_go.ServiceGroup{},
		EnvironmentVariables:map[string]string{"var1": "NALEJ_SERV_SERVICE1:2000", "var2": "NALEJ_SERV_SERVICE2"},
	}
}

func CreateAppDescriptorWithRepeatedGroup()* grpc_application_go.AddAppDescriptorRequest{

	return &grpc_application_go.AddAppDescriptorRequest {
		RequestId: 		uuid.New().String(),
		OrganizationId:	uuid.New().String(),
		Name: 			"descriptor-test",
		Rules: 			[]*grpc_application_go.SecurityRule{
			{
				Name: "rule1",
				TargetServiceGroupName: "g1",
				TargetServiceName:"service1",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g2",
				AuthServices: []string{"service3"},

			},
			{
				Name: "rule2",
				TargetServiceGroupName: "g2",
				TargetServiceName:"service3",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g1",
				AuthServices: []string{"service1", "service2"},
			},
		},
		Groups: 		[]*grpc_application_go.ServiceGroup{
			{
				Name:"g1",
				Services:[]*grpc_application_go.Service{
					{Name: "service1",},
					{Name: "service2",},},
			},
			{
				Name:"g1",
				Services:[]*grpc_application_go.Service{
					{Name: "service3",},},
			},
		},
	}

}

func CreateAppDescriptorWithRepeatedService()* grpc_application_go.AddAppDescriptorRequest{

	return &grpc_application_go.AddAppDescriptorRequest {
		RequestId: 		uuid.New().String(),
		OrganizationId:	uuid.New().String(),
		Name: 			"descriptor-test",
		Rules: 			[]*grpc_application_go.SecurityRule{
			{
				Name: "rule1",
				TargetServiceGroupName: "g1",
				TargetServiceName:"service1",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g2",
				AuthServices: []string{"service1"},

			},
			{
				Name: "rule2",
				TargetServiceGroupName: "g2",
				TargetServiceName:"service3",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g1",
				AuthServices: []string{"service1", "service2"},
			},
		},
		Groups: 		[]*grpc_application_go.ServiceGroup{
			{
				Name:"g1",
				Services:[]*grpc_application_go.Service{
					{Name: "service1",},
					{Name: "service1",},},
			},
			{
				Name:"g2",
				Services:[]*grpc_application_go.Service{
					{Name: "service3",},},
			},
		},
	}

}

func CreateAppDescriptorWrongGroupInRule()* grpc_application_go.AddAppDescriptorRequest{

	return &grpc_application_go.AddAppDescriptorRequest {
		RequestId: 		uuid.New().String(),
		OrganizationId:	uuid.New().String(),
		Name: 			"descriptor-test",
		Rules: 			[]*grpc_application_go.SecurityRule{
			{
				Name: "rule1",
				TargetServiceGroupName: "g7",
				TargetServiceName:"service1",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g2",
				AuthServices: []string{"service3"},

			},
			{
				Name: "rule2",
				TargetServiceGroupName: "g2",
				TargetServiceName:"service3",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g1",
				AuthServices: []string{"service1", "service2"},
			},
		},
		Groups: 		[]*grpc_application_go.ServiceGroup{
			{
				Name:"g1",
				Services:[]*grpc_application_go.Service{
					{Name: "service1",},
					{Name: "service2",},},
			},
			{
				Name:"g2",
				Services:[]*grpc_application_go.Service{
					{Name: "service3",},},
			},
		},
	}
}

func CreateAppDescriptorWrongDeployAfter()* grpc_application_go.AddAppDescriptorRequest{

	return &grpc_application_go.AddAppDescriptorRequest {
		RequestId: 		uuid.New().String(),
		OrganizationId:	uuid.New().String(),
		Name: 			"descriptor-test",
		Rules: 			[]*grpc_application_go.SecurityRule{
			{
				Name: "rule1",
				TargetServiceGroupName: "g1",
				TargetServiceName:"service1",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g2",
				AuthServices: []string{"service3"},

			},
			{
				Name: "rule2",
				TargetServiceGroupName: "g2",
				TargetServiceName:"service3",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g1",
				AuthServices: []string{"service1", "service2"},
			},
		},
		Groups: 		[]*grpc_application_go.ServiceGroup{
			{
				Name:"g1",
				Services:[]*grpc_application_go.Service{
					{Name: "service1", DeployAfter:[]string{"service2", "service5"}, },
					{Name: "service2",},},
			},
			{
				Name:"g2",
				Services:[]*grpc_application_go.Service{
					{Name: "service3",},},
			},
		},
	}
}

func CreateAppDescriptorWrongGroupDeploySpecs()* grpc_application_go.AddAppDescriptorRequest{

	return &grpc_application_go.AddAppDescriptorRequest {
		RequestId: 		uuid.New().String(),
		OrganizationId:	uuid.New().String(),
		Name: 			"descriptor-test",
		Rules: 			[]*grpc_application_go.SecurityRule{
			{
				Name: "rule1",
				TargetServiceGroupName: "g1",
				TargetServiceName:"service1",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g2",
				AuthServices: []string{"service3"},

			},
			{
				Name: "rule2",
				TargetServiceGroupName: "g2",
				TargetServiceName:"service3",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g1",
				AuthServices: []string{"service1", "service2"},
			},
		},
		Groups: 		[]*grpc_application_go.ServiceGroup{
			{
				Name:"g1",
				Services:[]*grpc_application_go.Service{
					{Name: "service1",},
					{Name: "service2",},},
				Specs: &grpc_application_go.ServiceGroupDeploymentSpecs{
					Replicas: 3,
					MultiClusterReplica: true,
				},
			},
			{
				Name:"g2",
				Services:[]*grpc_application_go.Service{
					{Name: "service3",},},
			},
		},
	}
}

func CreateAppDescriptorServiceToService()* grpc_application_go.AddAppDescriptorRequest{

	return &grpc_application_go.AddAppDescriptorRequest {
		RequestId: 		uuid.New().String(),
		OrganizationId:	uuid.New().String(),
		Name: 			"descriptor-test",
		Rules: 			[]*grpc_application_go.SecurityRule{
			{
				Name: "rule1",
				TargetServiceGroupName: "g1",
				TargetServiceName:"service1",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g2",
				AuthServices: []string{"service3"},

			},
			{
				Name: "rule2",
				TargetServiceGroupName: "g2",
				TargetServiceName:"service3",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g1",
				AuthServices: []string{"service1", "service2"},
			},
		},
		Groups: 		[]*grpc_application_go.ServiceGroup{
			{
				Name:"g1",
				Services:[]*grpc_application_go.Service{
					{Name: "service1",},
					{Name: "service2",},},
				Specs: &grpc_application_go.ServiceGroupDeploymentSpecs{
					Replicas: 3,
					MultiClusterReplica: false,
				},
			},
			{
				Name:"g2",
				Services:[]*grpc_application_go.Service{
					{Name: "service3",},},
			},
		},
	}
}

func CreateAppDescriptorWrongEnvironmentVariables()* grpc_application_go.AddAppDescriptorRequest{

	return &grpc_application_go.AddAppDescriptorRequest {
		RequestId: 		uuid.New().String(),
		OrganizationId:	uuid.New().String(),
		Name: 			"descriptor-test",
		Rules: 			[]*grpc_application_go.SecurityRule{
			{
				Name: "rule1",
				TargetServiceGroupName: "g1",
				TargetServiceName:"service1",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g2",
				AuthServices: []string{"service3"},

			},
			{
				Name: "rule2",
				TargetServiceGroupName: "g2",
				TargetServiceName:"service3",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g1",
				AuthServices: []string{"service1", "service2"},
			},
		},
		Groups: 		[]*grpc_application_go.ServiceGroup{
			{
				Name:"g1",
				Services:[]*grpc_application_go.Service{
					{Name: "service1",},
					{Name: "service2",},},
				Specs: &grpc_application_go.ServiceGroupDeploymentSpecs{
					Replicas: 3,
					MultiClusterReplica: false,
				},
			},
			{
				Name:"g2",
				Services:[]*grpc_application_go.Service{
					{Name: "service3",},},
			},
		},
		EnvironmentVariables:map[string]string{"var1": "NALEJ_SERV_SERVICE10:2000"},
	}
}

func CreateAppDescriptorWithDeviceRules()* grpc_application_go.AddAppDescriptorRequest{

	return &grpc_application_go.AddAppDescriptorRequest {
		RequestId: 		uuid.New().String(),
		OrganizationId:	uuid.New().String(),
		Name: 			"descriptor-test",
		Rules: 			[]*grpc_application_go.SecurityRule{
			{
				Name: "rule1",
				TargetServiceGroupName: "g1",
				TargetServiceName:"service1",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g2",
				AuthServices: []string{"service3"},

			},
			{
				Name: "rule2",
				TargetServiceGroupName: "g2",
				TargetServiceName:"service3",
				Access: grpc_application_go.PortAccess_DEVICE_GROUP,
				DeviceGroupNames:[]string{"deviceGroup1"},
			},
		},
		Groups: 		[]*grpc_application_go.ServiceGroup{
			{
				Name:"g1",
				Services:[]*grpc_application_go.Service{
					{Name: "service1",},
					{Name: "service2",},},
				Specs: &grpc_application_go.ServiceGroupDeploymentSpecs{
					Replicas: 3,
					MultiClusterReplica: false,
				},
			},
			{
				Name:"g2",
				Services:[]*grpc_application_go.Service{
					{Name: "service3",},},
			},
		},
		EnvironmentVariables:map[string]string{"var1": "NALEJ_SERV_SERVICE1:2000", "var2": "NALEJ_SERV_SERVICE2"},
	}
}
func CreateAppDescriptorWithWrongDeviceRules()* grpc_application_go.AddAppDescriptorRequest{

	return &grpc_application_go.AddAppDescriptorRequest {
		RequestId: 		uuid.New().String(),
		OrganizationId:	uuid.New().String(),
		Name: 			"descriptor-test",
		Rules: 			[]*grpc_application_go.SecurityRule{
			{
				Name: "rule1",
				TargetServiceGroupName: "g1",
				TargetServiceName:"service1",
				Access: grpc_application_go.PortAccess_APP_SERVICES,
				AuthServiceGroupName: "g2",
				AuthServices: []string{"service3"},

			},
			{
				Name: "rule2",
				TargetServiceGroupName: "g2",
				TargetServiceName:"service3",
				Access: grpc_application_go.PortAccess_DEVICE_GROUP,
				AuthServiceGroupName: "g1",
				AuthServices: []string{"service1", "service2"},
				DeviceGroupNames:[]string{"deviceGroup1"},
			},
		},
		Groups: 		[]*grpc_application_go.ServiceGroup{
			{
				Name:"g1",
				Services:[]*grpc_application_go.Service{
					{Name: "service1",},
					{Name: "service2",},},
				Specs: &grpc_application_go.ServiceGroupDeploymentSpecs{
					Replicas: 3,
					MultiClusterReplica: false,
				},
			},
			{
				Name:"g2",
				Services:[]*grpc_application_go.Service{
					{Name: "service3",},},
			},
		},
		EnvironmentVariables:map[string]string{"var1": "NALEJ_SERV_SERVICE1:2000", "var2": "NALEJ_SERV_SERVICE2"},
	}
}