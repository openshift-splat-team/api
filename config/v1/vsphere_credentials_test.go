package v1

import (
	"encoding/json"
	"testing"
)

// AC1: VSphereCredentialsMode zero-value defaults to Passthrough

func TestVSphereCredentialsMode_ZeroValue(t *testing.T) {
	var mode VSphereCredentialsMode
	if mode != VSphereCredentialsModePassthrough {
		t.Errorf("zero value of VSphereCredentialsMode = %q; want %q (Passthrough)",
			mode, VSphereCredentialsModePassthrough)
	}
}

func TestVSpherePlatformSpec_CredentialsModeUnset_DefaultsPassthrough(t *testing.T) {
	spec := VSpherePlatformSpec{}
	if spec.CredentialsMode != VSphereCredentialsModePassthrough {
		t.Errorf("unset CredentialsMode = %q; want Passthrough", spec.CredentialsMode)
	}
}

func TestVSpherePlatformSpec_CredentialsModePassthrough(t *testing.T) {
	spec := VSpherePlatformSpec{
		CredentialsMode: VSphereCredentialsModePassthrough,
	}
	if spec.CredentialsMode != VSphereCredentialsModePassthrough {
		t.Errorf("CredentialsMode = %q; want Passthrough", spec.CredentialsMode)
	}
}

func TestVSpherePlatformSpec_CredentialsModePerComponent(t *testing.T) {
	spec := VSpherePlatformSpec{
		CredentialsMode: VSphereCredentialsModePerComponent,
	}
	if spec.CredentialsMode != VSphereCredentialsModePerComponent {
		t.Errorf("CredentialsMode = %q; want PerComponent", spec.CredentialsMode)
	}
}

// AC2: Existing cluster without credentialsMode → Passthrough, no regression

func TestVSpherePlatformSpec_JSONRoundTrip_NoCredentialsMode(t *testing.T) {
	raw := `{"vcenters": [{"server": "vcenter.example.com", "port": 443, "datacenters": ["DC1"]}]}`

	var spec VSpherePlatformSpec
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if spec.CredentialsMode != VSphereCredentialsModePassthrough {
		t.Errorf("credentialsMode after unmarshal of legacy JSON = %q; want Passthrough", spec.CredentialsMode)
	}
	if spec.ComponentCredentials != nil {
		t.Errorf("ComponentCredentials should be nil for legacy JSON; got %+v", spec.ComponentCredentials)
	}
}

func TestVSpherePlatformSpec_JSONRoundTrip_ExplicitPassthrough(t *testing.T) {
	// VSphereCredentialsModePassthrough == "" so the canonical representation is omitting the field.
	// An explicitly empty credentialsMode field is equally valid and must round-trip as Passthrough.
	raw := `{"credentialsMode": ""}`

	var spec VSpherePlatformSpec
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if spec.CredentialsMode != VSphereCredentialsModePassthrough {
		t.Errorf("credentialsMode = %q; want Passthrough (empty string)", spec.CredentialsMode)
	}
}

// AC3: PerComponent mode → component credentials accessible

func TestVSpherePlatformSpec_PerComponent_AllComponents(t *testing.T) {
	spec := VSpherePlatformSpec{
		CredentialsMode: VSphereCredentialsModePerComponent,
		ComponentCredentials: &VSphereComponentCredentials{
			MachineAPI: &VSphereComponentSecretRef{
				Name:      "vsphere-machine-api-creds",
				Namespace: "openshift-machine-api",
			},
			CSIDriver: &VSphereComponentSecretRef{
				Name:      "vsphere-storage-creds",
				Namespace: "openshift-cluster-csi-drivers",
			},
			CloudController: &VSphereComponentSecretRef{
				Name:      "vsphere-cloud-controller-creds",
				Namespace: "openshift-cloud-controller-manager",
			},
			VSphereProblemDetector: &VSphereComponentSecretRef{
				Name:      "vsphere-problem-detector-creds",
				Namespace: "openshift-cluster-storage-operator",
			},
		},
	}

	cc := spec.ComponentCredentials
	if cc == nil {
		t.Fatal("ComponentCredentials is nil; want non-nil")
	}
	if cc.MachineAPI == nil || cc.MachineAPI.Name != "vsphere-machine-api-creds" {
		t.Error("MachineAPI credential missing or wrong name")
	}
	if cc.CSIDriver == nil || cc.CSIDriver.Name != "vsphere-storage-creds" {
		t.Error("CSIDriver credential missing or wrong name")
	}
	if cc.CloudController == nil || cc.CloudController.Name != "vsphere-cloud-controller-creds" {
		t.Error("CloudController credential missing or wrong name")
	}
	if cc.VSphereProblemDetector == nil || cc.VSphereProblemDetector.Name != "vsphere-problem-detector-creds" {
		t.Error("VSphereProblemDetector credential missing or wrong name")
	}
}

func TestVSpherePlatformSpec_PerComponent_PartialComponents(t *testing.T) {
	spec := VSpherePlatformSpec{
		CredentialsMode: VSphereCredentialsModePerComponent,
		ComponentCredentials: &VSphereComponentCredentials{
			MachineAPI: &VSphereComponentSecretRef{
				Name:      "vsphere-machine-api-creds",
				Namespace: "openshift-machine-api",
			},
		},
	}
	cc := spec.ComponentCredentials
	if cc.MachineAPI == nil {
		t.Error("MachineAPI should be non-nil")
	}
	if cc.CSIDriver != nil {
		t.Errorf("CSIDriver should be nil; got %+v", cc.CSIDriver)
	}
	if cc.CloudController != nil {
		t.Errorf("CloudController should be nil; got %+v", cc.CloudController)
	}
	if cc.VSphereProblemDetector != nil {
		t.Errorf("VSphereProblemDetector should be nil; got %+v", cc.VSphereProblemDetector)
	}
}

func TestVSpherePlatformSpec_PerComponent_JSONRoundTrip(t *testing.T) {
	original := VSpherePlatformSpec{
		CredentialsMode: VSphereCredentialsModePerComponent,
		ComponentCredentials: &VSphereComponentCredentials{
			MachineAPI: &VSphereComponentSecretRef{
				Name:      "vsphere-machine-api-creds",
				Namespace: "openshift-machine-api",
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var roundTripped VSpherePlatformSpec
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if roundTripped.CredentialsMode != VSphereCredentialsModePerComponent {
		t.Errorf("credentialsMode after round-trip = %q; want PerComponent", roundTripped.CredentialsMode)
	}
	if roundTripped.ComponentCredentials == nil || roundTripped.ComponentCredentials.MachineAPI == nil {
		t.Error("MachineAPI not preserved in round-trip")
	}
	if roundTripped.ComponentCredentials.MachineAPI.Name != "vsphere-machine-api-creds" {
		t.Errorf("Name after round-trip = %q; want vsphere-machine-api-creds",
			roundTripped.ComponentCredentials.MachineAPI.Name)
	}
}

// Adversarial cases

func TestVSpherePlatformSpec_EmptyStringCredentialsMode(t *testing.T) {
	raw := `{"credentialsMode": ""}`
	var spec VSpherePlatformSpec
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if spec.CredentialsMode != VSphereCredentialsModePassthrough {
		t.Errorf("empty credentialsMode = %q; want Passthrough", spec.CredentialsMode)
	}
}

func TestVSpherePlatformSpec_ComponentCredentials_NilPointerSafety(t *testing.T) {
	spec := VSpherePlatformSpec{
		CredentialsMode:      VSphereCredentialsModePassthrough,
		ComponentCredentials: nil,
	}
	if spec.ComponentCredentials != nil {
		t.Fatalf("expected nil ComponentCredentials")
	}
	var machineAPICreds *VSphereComponentSecretRef
	if spec.ComponentCredentials != nil {
		machineAPICreds = spec.ComponentCredentials.MachineAPI
	}
	_ = machineAPICreds
}

func TestVSphereComponentSecretRef_EmptyNamespace(t *testing.T) {
	ref := VSphereComponentSecretRef{Name: "some-secret", Namespace: ""}
	if ref.Name != "some-secret" {
		t.Errorf("Name = %q; want some-secret", ref.Name)
	}
}

func TestVSphereComponentSecretRef_EmptyName(t *testing.T) {
	ref := VSphereComponentSecretRef{Name: "", Namespace: "openshift-machine-api"}
	if ref.Namespace != "openshift-machine-api" {
		t.Errorf("Namespace = %q; want openshift-machine-api", ref.Namespace)
	}
}

func TestVSpherePlatformSpec_DeepCopy_PointerIsolation(t *testing.T) {
	original := &VSpherePlatformSpec{
		CredentialsMode: VSphereCredentialsModePerComponent,
		ComponentCredentials: &VSphereComponentCredentials{
			MachineAPI: &VSphereComponentSecretRef{
				Name:      "vsphere-machine-api-creds",
				Namespace: "openshift-machine-api",
			},
		},
	}
	copied := original.DeepCopy()
	original.ComponentCredentials.MachineAPI.Name = "mutated"
	if copied.ComponentCredentials.MachineAPI.Name == "mutated" {
		t.Error("DeepCopy shares pointer with original: mutation of original affected copy")
	}
}
