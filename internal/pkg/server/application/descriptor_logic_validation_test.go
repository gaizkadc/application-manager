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
	"github.com/nalej/application-manager/internal/pkg/entities"
	"github.com/nalej/application-manager/internal/pkg/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Application Descriptor Validations", func() {

	ginkgo.Context("Valid app descriptors", func() {
		ginkgo.It("should pass the validation", func() {
			appDescriptor := utils.CreateFullAppDescriptor()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).To(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (repeated group)", func() {
			appDescriptor := utils.CreateAppDescriptorWithRepeatedGroup()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (repeated service)", func() {
			appDescriptor := utils.CreateAppDescriptorWithRepeatedService()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (wrong group in rule)", func() {
			appDescriptor := utils.CreateAppDescriptorWrongGroupInRule()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (wrong deploy after service)", func() {
			appDescriptor := utils.CreateAppDescriptorWrongDeployAfter()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (wrong group deploy specs)", func() {
			appDescriptor := utils.CreateAppDescriptorWrongGroupDeploySpecs()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.PIt("should not pass the validation (service to service rule)", func() {
			appDescriptor := utils.CreateAppDescriptorServiceToService()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (wrong environment variables)", func() {
			appDescriptor := utils.CreateAppDescriptorWrongEnvironmentVariables()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should pass the validation (device group access)", func() {
			appDescriptor := utils.CreateAppDescriptorWithDeviceRules()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).To(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (wrong device group access)", func() {
			appDescriptor := utils.CreateAppDescriptorWithWrongDeviceRules()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should not pass the validation (app descriptor without groups)", func() {
			appDescriptor := utils.CreateAppDescriptorWithoutGroups()
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})
		ginkgo.It("should pass the validation", func() {
			appDescriptor := utils.CreateFullAppDescriptor()
			appDescriptor.EnvironmentVariables["sonar.jdbc.username=sonar"] = "sonar"
			err := entities.ValidDescriptorLogic(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})

	})

	ginkgo.Context("Valid storage path", func() {
		ginkgo.It("should pass the validation ", func() {
			appDescriptor := utils.CreateTestAddDescriptorWithMountPath()
			err := entities.ValidateStoragePathAppRequest(appDescriptor)
			gomega.Expect(err).To(gomega.Succeed())
		})
		ginkgo.It("should not pass the storage validation ", func() {
			appDescriptor := utils.CreateTestAddDescriptorWithWrongMountPath()
			err := entities.ValidateStoragePathAppRequest(appDescriptor)
			gomega.Expect(err).NotTo(gomega.Succeed())
		})

	})
})
