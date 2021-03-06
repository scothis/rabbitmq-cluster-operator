// RabbitMQ Cluster Operator
//
// Copyright 2020 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Mozilla Public license, Version 2.0 (the "License").  You may not use this product except in compliance with the Mozilla Public License.
//
// This product may include a number of subcomponents with separate copyright notices and license terms. Your use of these subcomponents is subject to the terms and conditions of the subcomponent's license, as noted in the LICENSE file.
//

package status_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rabbitmqstatus "github.com/pivotal/rabbitmq-for-kubernetes/internal/status"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("AllReplicasReady", func() {
	var (
		sts               *appsv1.StatefulSet
		existingCondition *rabbitmqstatus.RabbitmqClusterCondition
	)

	BeforeEach(func() {
		sts = &appsv1.StatefulSet{
			Status: appsv1.StatefulSetStatus{},
		}
		existingCondition = nil

	})

	Context("condition status and reason", func() {
		When("all replicas are ready", func() {
			BeforeEach(func() {
				sts.Status.ReadyReplicas = 5
				sts.Status.Replicas = 5
			})

			It("returns the expected condition", func() {
				condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{&corev1.Endpoints{}, sts}, existingCondition)

				By("having status true and reason message", func() {
					Expect(condition.Status).To(Equal(corev1.ConditionTrue))
					Expect(condition.Reason).To(Equal("AllPodsAreReady"))
				})
			})
		})

		When("some replicas are not ready", func() {
			BeforeEach(func() {
				sts.Status.ReadyReplicas = 3
				sts.Status.Replicas = 5
			})

			It("returns a condition with state false", func() {
				condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)

				By("having status false and reason", func() {
					Expect(condition.Status).To(Equal(corev1.ConditionFalse))
					Expect(condition.Reason).To(Equal("NotAllPodsReady"))
					Expect(condition.Message).ToNot(BeEmpty())
				})
			})
		})

		When("the StatefulSet is not found", func() {
			BeforeEach(func() {
				sts = nil
			})

			It("returns a condition with state unknown", func() {
				condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)

				By("having status unknown and reason", func() {
					Expect(condition.Status).To(Equal(corev1.ConditionUnknown))
					Expect(condition.Reason).To(Equal("MissingStatefulSet"))
					Expect(condition.Message).ToNot(BeEmpty())
				})
			})
		})
	})

	Context("Condition transitions", func() {
		var (
			previousConditionTime time.Time
		)

		BeforeEach(func() {
			previousConditionTime = time.Date(2020, 2, 2, 8, 0, 0, 0, time.UTC)
		})

		Context("previous condition was not set", func() {
			var (
				emptyTime metav1.Time
			)

			BeforeEach(func() {
				emptyTime = metav1.Time{}
			})

			When("transitions to true", func() {
				BeforeEach(func() {
					sts.Status.ReadyReplicas = 5
					sts.Status.Replicas = 5
				})

				It("updates the transition time", func() {
					condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{&corev1.Endpoints{}, sts}, existingCondition)
					Expect(condition.LastTransitionTime).ToNot(Equal(emptyTime))
				})
			})

			When("transitions to false", func() {
				BeforeEach(func() {
					sts.Status.ReadyReplicas = 3
					sts.Status.Replicas = 5
				})

				It("updates the transition time", func() {
					condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)
					Expect(condition.LastTransitionTime).ToNot(Equal(emptyTime))
				})
			})

			When("transitions to unknown", func() {
				BeforeEach(func() {
					sts = nil
				})

				It("updates the transition time", func() {
					condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)
					Expect(condition.LastTransitionTime).ToNot(Equal(emptyTime))
				})
			})
		})

		Context("previous condition was false", func() {
			BeforeEach(func() {
				existingCondition = &rabbitmqstatus.RabbitmqClusterCondition{
					Status: corev1.ConditionFalse,
					LastTransitionTime: metav1.Time{
						Time: previousConditionTime,
					},
				}
			})

			When("transitions to true", func() {
				BeforeEach(func() {
					sts.Status.ReadyReplicas = 5
					sts.Status.Replicas = 5
				})

				It("updates the transition time", func() {
					condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)

					Expect(existingCondition).NotTo(BeNil())
					existingConditionTime := existingCondition.LastTransitionTime.DeepCopy()

					Expect(condition.LastTransitionTime.Equal(existingConditionTime)).To(BeFalse())
					Expect(condition.LastTransitionTime.Before(existingConditionTime)).To(BeFalse())
				})
			})

			When("remains false", func() {
				BeforeEach(func() {
					sts.Status.ReadyReplicas = 3
					sts.Status.Replicas = 5
				})

				It("does not update transition time", func() {
					condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)

					Expect(existingCondition).NotTo(BeNil())
					existingConditionTime := existingCondition.LastTransitionTime.DeepCopy()
					Expect(condition.LastTransitionTime.Equal(existingConditionTime)).To(BeTrue())
				})
			})

			When("transitions to unknown", func() {
				BeforeEach(func() {
					sts = nil
				})

				It("updates the transition time", func() {
					condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)

					Expect(existingCondition).NotTo(BeNil())
					existingConditionTime := existingCondition.LastTransitionTime.DeepCopy()
					Expect(condition.LastTransitionTime.Equal(existingConditionTime)).To(BeFalse())
					Expect(condition.LastTransitionTime.Before(existingConditionTime)).To(BeFalse())
				})
			})
		})

		Context("previous condition was true", func() {
			BeforeEach(func() {
				existingCondition = &rabbitmqstatus.RabbitmqClusterCondition{
					Status: corev1.ConditionTrue,
					LastTransitionTime: metav1.Time{
						Time: previousConditionTime,
					},
				}
			})

			When("remains true", func() {
				BeforeEach(func() {
					sts.Status.ReadyReplicas = 5
					sts.Status.Replicas = 5
				})

				It("does not update transition time", func() {
					condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)

					Expect(existingCondition).NotTo(BeNil())
					existingConditionTime := existingCondition.LastTransitionTime.DeepCopy()
					Expect(condition.LastTransitionTime.Equal(existingConditionTime)).To(BeTrue())
				})
			})

			When("transitions to false", func() {
				BeforeEach(func() {
					sts.Status.ReadyReplicas = 3
					sts.Status.Replicas = 5
				})

				It("updates the transition time", func() {
					condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)

					Expect(existingCondition).NotTo(BeNil())
					existingConditionTime := existingCondition.LastTransitionTime.DeepCopy()
					Expect(condition.LastTransitionTime.Equal(existingConditionTime)).To(BeFalse())
					Expect(condition.LastTransitionTime.Before(existingConditionTime)).To(BeFalse())
				})
			})

			When("transitions to unknown", func() {
				BeforeEach(func() {
					sts = nil
				})

				It("updates the transition time", func() {
					condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)

					Expect(existingCondition).NotTo(BeNil())
					existingConditionTime := existingCondition.LastTransitionTime.DeepCopy()
					Expect(condition.LastTransitionTime.Equal(existingConditionTime)).To(BeFalse())
					Expect(condition.LastTransitionTime.Before(existingConditionTime)).To(BeFalse())
				})
			})
		})

		Context("previous condition was unknown", func() {
			BeforeEach(func() {
				existingCondition = &rabbitmqstatus.RabbitmqClusterCondition{
					Status: corev1.ConditionUnknown,
					LastTransitionTime: metav1.Time{
						Time: previousConditionTime,
					},
				}
			})

			When("transitions to true", func() {
				BeforeEach(func() {
					sts.Status.ReadyReplicas = 5
					sts.Status.Replicas = 5
				})

				It("updates the transition time", func() {
					condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)

					Expect(existingCondition).NotTo(BeNil())
					existingConditionTime := existingCondition.LastTransitionTime.DeepCopy()
					Expect(condition.LastTransitionTime.Equal(existingConditionTime)).To(BeFalse())
					Expect(condition.LastTransitionTime.Before(existingConditionTime)).To(BeFalse())
				})
			})

			When("transitions to false", func() {
				BeforeEach(func() {
					sts.Status.ReadyReplicas = 3
					sts.Status.Replicas = 5
				})

				It("updates the transition time", func() {
					condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)

					Expect(existingCondition).NotTo(BeNil())
					existingConditionTime := existingCondition.LastTransitionTime.DeepCopy()
					Expect(condition.LastTransitionTime.Equal(existingConditionTime)).To(BeFalse())
					Expect(condition.LastTransitionTime.Before(existingConditionTime)).To(BeFalse())
				})
			})

			When("remains unknown", func() {
				BeforeEach(func() {
					sts = nil
				})

				It("does not update transition time", func() {
					condition := rabbitmqstatus.AllReplicasReadyCondition([]runtime.Object{sts}, existingCondition)

					Expect(existingCondition).NotTo(BeNil())
					existingConditionTime := existingCondition.LastTransitionTime.DeepCopy()
					Expect(condition.LastTransitionTime.Equal(existingConditionTime)).To(BeTrue())
				})
			})
		})
	})
})
