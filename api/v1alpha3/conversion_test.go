/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha3

import (
	"testing"

	fuzz "github.com/google/gofuzz"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"k8s.io/apimachinery/pkg/runtime"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/cluster-api/api/v1alpha4"
	utilconversion "sigs.k8s.io/cluster-api/util/conversion"
)

func TestFuzzyConversion(t *testing.T) {
	g := NewWithT(t)
	scheme := runtime.NewScheme()
	g.Expect(AddToScheme(scheme)).To(Succeed())
	g.Expect(v1alpha4.AddToScheme(scheme)).To(Succeed())

	t.Run("for Cluster", utilconversion.FuzzTestFunc(utilconversion.FuzzTestFuncInput{
		Scheme:             scheme,
		Hub:                &v1alpha4.Cluster{},
		Spoke:              &Cluster{},
		SpokeAfterMutation: clusterSpokeAfterMutation,
	}))

	t.Run("for Machine", utilconversion.FuzzTestFunc(utilconversion.FuzzTestFuncInput{
		Scheme:      scheme,
		Hub:         &v1alpha4.Machine{},
		Spoke:       &Machine{},
		FuzzerFuncs: []fuzzer.FuzzerFuncs{BootstrapFuzzFuncs},
	}))

	t.Run("for MachineSet", utilconversion.FuzzTestFunc(utilconversion.FuzzTestFuncInput{
		Scheme:      scheme,
		Hub:         &v1alpha4.MachineSet{},
		Spoke:       &MachineSet{},
		FuzzerFuncs: []fuzzer.FuzzerFuncs{BootstrapFuzzFuncs, CustomObjectMetaFuzzFunc},
	}))

	t.Run("for MachineDeployment", utilconversion.FuzzTestFunc(utilconversion.FuzzTestFuncInput{
		Scheme:      scheme,
		Hub:         &v1alpha4.MachineDeployment{},
		Spoke:       &MachineDeployment{},
		FuzzerFuncs: []fuzzer.FuzzerFuncs{BootstrapFuzzFuncs, CustomObjectMetaFuzzFunc},
	}))

	t.Run("for MachineHealthCheckSpec", utilconversion.FuzzTestFunc(utilconversion.FuzzTestFuncInput{
		Scheme: scheme,
		Hub:    &v1alpha4.MachineHealthCheck{},
		Spoke:  &MachineHealthCheck{},
	}))
}

func CustomObjectMetaFuzzFunc(_ runtimeserializer.CodecFactory) []interface{} {
	return []interface{}{
		CustomObjectMetaFuzzer,
	}
}

func CustomObjectMetaFuzzer(in *ObjectMeta, c fuzz.Continue) {
	c.FuzzNoCustom(in)

	// These fields have been removed in v1alpha4
	// data is going to be lost, so we're forcing zero values here.
	in.Name = ""
	in.GenerateName = ""
	in.Namespace = ""
	in.OwnerReferences = nil
}

func BootstrapFuzzFuncs(_ runtimeserializer.CodecFactory) []interface{} {
	return []interface{}{
		BootstrapFuzzer,
	}
}

func BootstrapFuzzer(obj *Bootstrap, c fuzz.Continue) {
	c.FuzzNoCustom(obj)

	// Bootstrap.Data has been removed in v1alpha4, so setting it to nil in order to avoid v1alpha3 --> v1alpha4 --> v1alpha3 round trip errors.
	obj.Data = nil
}

// clusterSpokeAfterMutation modifies the spoke version of the Cluster such that it can pass an equality test in the
// spoke-hub-spoke conversion scenario.
func clusterSpokeAfterMutation(c conversion.Convertible) {
	cluster := c.(*Cluster)

	// Create a temporary 0-length slice using the same underlying array as cluster.Status.Conditions to avoid
	// allocations.
	tmp := cluster.Status.Conditions[:0]

	for i := range cluster.Status.Conditions {
		condition := cluster.Status.Conditions[i]

		// Keep everything that is not ControlPlaneInitializedCondition
		if condition.Type != ConditionType(v1alpha4.ControlPlaneInitializedCondition) {
			tmp = append(tmp, condition)
		}
	}

	// Point cluster.Status.Conditions and our slice that does not have ControlPlaneInitializedCondition
	cluster.Status.Conditions = tmp
}
