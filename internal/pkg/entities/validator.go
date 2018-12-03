package entities

import (
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-organization-go"
)

const emptyRequestId = "request_id cannot be empty"
const emptyOrganizationId = "organization_id cannot be empty"
const emptyDescriptorId = "app_descriptor_id cannot be empty"
const emptyInstanceId = "app_instance_id cannot be empty"
const emptyName = "name cannot be empty"


func ValidOrganizationId(organizationID *grpc_organization_go.OrganizationId) derrors.Error {
	if organizationID.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	return nil
}

func ValidAddAppDescriptorRequest(toAdd * grpc_application_go.AddAppDescriptorRequest) derrors.Error {
	if toAdd.OrganizationId == ""{
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if toAdd.RequestId == "" {
		return derrors.NewInvalidArgumentError(emptyRequestId)
	}

	if toAdd.Name == ""{
		return derrors.NewInvalidArgumentError(emptyName)
	}

	if len(toAdd.Services) == 0 {
		return derrors.NewInvalidArgumentError("expecting at least one service")
	}

	return nil
}

func ValidAppDescriptorID(descriptorID * grpc_application_go.AppDescriptorId) derrors.Error {
	if descriptorID.OrganizationId == ""{
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if descriptorID.AppDescriptorId == "" {
		return derrors.NewInvalidArgumentError(emptyDescriptorId)
	}
	return nil
}

func ValidAppInstanceID(instanceID * grpc_application_go.AppInstanceId) derrors.Error {
	if instanceID.OrganizationId == ""{
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if instanceID.AppInstanceId == "" {
		return derrors.NewInvalidArgumentError(emptyInstanceId)
	}
	return nil
}

func ValidDeployRequest(deployRequest *grpc_application_manager_go.DeployRequest) derrors.Error {
	if deployRequest.OrganizationId == ""{
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}

	if deployRequest.AppDescriptorId == ""{
		return derrors.NewInvalidArgumentError(emptyDescriptorId)
	}

	if deployRequest.Name == ""{
		return derrors.NewInvalidArgumentError(emptyName)
	}

	return nil
}