package providers

import "testing"

func TestCapabilitiesForKind(t *testing.T) {
	gh := CapabilitiesForKind(KindGitHub)
	if !gh.SupportsChecks || !gh.SupportsFiles || !gh.SupportsUpdateBranch {
		t.Fatalf("expected github to support checks/files/update branch")
	}
	gl := CapabilitiesForKind(KindGitLab)
	if gl.SupportsChecks || gl.SupportsFiles || gl.SupportsUpdateBranch {
		t.Fatalf("expected gitlab to not support checks/files/update branch")
	}
	if !gl.SupportsApprovals || !gl.SupportsMerge {
		t.Fatalf("expected gitlab to support approvals/merge")
	}
}
