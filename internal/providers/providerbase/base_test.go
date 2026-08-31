package providerbase

import (
	"net/http"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestNew_AppliesAPIKeyAuthDefaults(t *testing.T) {
	base := New(core.ProviderSpec{
		ID: "sample",
		Info: core.ProviderInfo{
			Name:   "Sample",
			DocURL: "https://example.com/docs",
		},
		Auth: core.ProviderAuthSpec{
			Type:      core.ProviderAuthTypeAPIKey,
			APIKeyEnv: "SAMPLE_API_KEY",
		},
	})

	spec := base.Spec()
	if spec.Setup.DocsURL != "https://example.com/docs" {
		t.Fatalf("setup docs = %q, want %q", spec.Setup.DocsURL, "https://example.com/docs")
	}
	// Auth metadata lives in Spec().Auth, not in DashboardWidget.
	if got := spec.Auth.APIKeyEnv; got != "SAMPLE_API_KEY" {
		t.Fatalf("spec auth APIKeyEnv = %q, want %q", got, "SAMPLE_API_KEY")
	}
	// DashboardWidget should NOT have auth fields copied into it.
	if got := base.DashboardWidget().APIKeyEnv; got != "" {
		t.Fatalf("dashboard APIKeyEnv = %q, want empty (auth data should not be copied)", got)
	}
	if got := base.DashboardWidget().DefaultAccountID; got != "" {
		t.Fatalf("dashboard DefaultAccountID = %q, want empty (auth data should not be copied)", got)
	}
	if len(base.DetailWidget().Sections) == 0 {
		t.Fatal("detail sections should have default sections")
	}
}

func TestNew_AuthMetadataInSpecNotWidget(t *testing.T) {
	base := New(core.ProviderSpec{
		ID: "sample",
		Info: core.ProviderInfo{
			Name: "Sample",
		},
		Auth: core.ProviderAuthSpec{
			Type:             core.ProviderAuthTypeAPIKey,
			APIKeyEnv:        "SAMPLE_API_KEY",
			DefaultAccountID: "sample-auth",
		},
		Dashboard: core.DashboardWidget{
			// Legacy fields still exist on the struct but should not be
			// the source of truth. TUI reads from Spec().Auth instead.
			APIKeyEnv:        "CUSTOM_API_KEY",
			DefaultAccountID: "sample-widget",
		},
		Detail: core.DefaultDetailWidget(),
	})

	spec := base.Spec()
	// The canonical source for auth metadata is Spec().Auth.
	if got := spec.Auth.APIKeyEnv; got != "SAMPLE_API_KEY" {
		t.Fatalf("spec auth APIKeyEnv = %q, want %q", got, "SAMPLE_API_KEY")
	}
	if got := spec.Auth.DefaultAccountID; got != "sample-auth" {
		t.Fatalf("spec auth DefaultAccountID = %q, want %q", got, "sample-auth")
	}
	// DashboardWidget preserves whatever was set explicitly on spec.Dashboard
	// (no copy logic from Auth), but TUI code should not read auth from here.
	if got := base.DashboardWidget().APIKeyEnv; got != "CUSTOM_API_KEY" {
		t.Fatalf("dashboard APIKeyEnv = %q, want %q (explicit value preserved)", got, "CUSTOM_API_KEY")
	}
	if spec.Dashboard.APIKeyEnv != "CUSTOM_API_KEY" {
		t.Fatalf("spec dashboard APIKeyEnv = %q, want %q", spec.Dashboard.APIKeyEnv, "CUSTOM_API_KEY")
	}
}

func TestNew_DefaultsAndNormalization(t *testing.T) {
	tests := []struct {
		name        string
		spec        core.ProviderSpec
		wantID      string
		wantName    string
		wantDocsURL string
	}{
		{
			name:        "empty spec gets unknown id and defaults",
			spec:        core.ProviderSpec{},
			wantID:      "unknown",
			wantName:    "unknown",
			wantDocsURL: "",
		},
		{
			name: "spec with id only sets name to id",
			spec: core.ProviderSpec{
				ID: "custom-id",
			},
			wantID:      "custom-id",
			wantName:    "custom-id",
			wantDocsURL: "",
		},
		{
			name: "spec with info doc url populates setup docs url when empty",
			spec: core.ProviderSpec{
				ID: "test-prov",
				Info: core.ProviderInfo{
					Name:   "Test Provider",
					DocURL: "https://example.com/info-doc",
				},
			},
			wantID:      "test-prov",
			wantName:    "Test Provider",
			wantDocsURL: "https://example.com/info-doc",
		},
		{
			name: "explicit setup docs url takes precedence over info doc url",
			spec: core.ProviderSpec{
				ID: "test-prov",
				Info: core.ProviderInfo{
					Name:   "Test Provider",
					DocURL: "https://example.com/info-doc",
				},
				Setup: core.ProviderSetupSpec{
					DocsURL: "https://example.com/setup-doc",
				},
			},
			wantID:      "test-prov",
			wantName:    "Test Provider",
			wantDocsURL: "https://example.com/setup-doc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.spec)
			if got := b.ID(); got != tt.wantID {
				t.Errorf("ID() = %q, want %q", got, tt.wantID)
			}
			if got := b.Describe().Name; got != tt.wantName {
				t.Errorf("Describe().Name = %q, want %q", got, tt.wantName)
			}
			if got := b.Spec().Setup.DocsURL; got != tt.wantDocsURL {
				t.Errorf("Spec().Setup.DocsURL = %q, want %q", got, tt.wantDocsURL)
			}
		})
	}
}

func TestBase_Client(t *testing.T) {
	// 1. Default client when HTTPClient is nil
	bDefault := Base{}
	clientDefault := bDefault.Client()
	if clientDefault == nil {
		t.Fatal("Client() returned nil for default Base")
	}
	if clientDefault.Timeout != 30*time.Second {
		t.Errorf("default client timeout = %v, want %v", clientDefault.Timeout, 30*time.Second)
	}

	// 2. Custom client when explicitly set
	customClient := &http.Client{Timeout: 5 * time.Second}
	bCustom := Base{HTTPClient: customClient}
	if got := bCustom.Client(); got != customClient {
		t.Errorf("Client() = %v, want custom client %v", got, customClient)
	}
}

func TestBase_Widgets(t *testing.T) {
	// Default zero dashboard fallback
	bDefault := New(core.ProviderSpec{ID: "test"})
	dash := bDefault.DashboardWidget()
	if dash.IsZero() {
		t.Error("DashboardWidget() on zero spec should return DefaultDashboardWidget(), but got zero widget")
	}

	// Custom dashboard preserved
	customDash := core.DashboardWidget{
		ShowToolComposition: true,
		ColorRole:           core.DashboardColorRoleTeal,
	}
	bCustomDash := New(core.ProviderSpec{
		ID:        "test",
		Dashboard: customDash,
	})
	if !bCustomDash.DashboardWidget().ShowToolComposition {
		t.Error("expected custom DashboardWidget to preserve ShowToolComposition = true")
	}

	// Default detail fallback when sections empty
	detail := bDefault.DetailWidget()
	if len(detail.Sections) == 0 {
		t.Error("DetailWidget() should return DefaultDetailWidget() when sections empty")
	}

	// Custom detail preserved
	customDetail := core.DetailWidget{
		Sections: []core.DetailSection{
			{Name: "Custom Section", Order: 1, Style: core.DetailSectionStyleUsage},
		},
	}
	bCustomDetail := New(core.ProviderSpec{
		ID:     "test",
		Detail: customDetail,
	})
	if len(bCustomDetail.DetailWidget().Sections) != 1 || bCustomDetail.DetailWidget().Sections[0].Name != "Custom Section" {
		t.Errorf("DetailWidget() did not preserve custom sections: %+v", bCustomDetail.DetailWidget().Sections)
	}
}

func TestDashboardOptions_AllMutators(t *testing.T) {
	labels := map[string]string{"cost": "Estimated Cost", "tokens": "Total Tokens"}
	compactLabels := map[string]string{"cost": "$", "tokens": "tok"}
	rawGroup := core.DashboardRawGroup{Label: "Raw Dump", Keys: []string{"raw1", "raw2"}}
	compactRow := core.DashboardCompactRow{Label: "Status", Keys: []string{"stat"}}

	dash := DefaultDashboard(
		nil, // Nil safety check
		WithColorRole(core.DashboardColorRoleLavender),
		WithGaugePriority("quota_weekly", "quota_5h"),
		WithGaugeMaxLines(4),
		WithCompactRows(compactRow),
		WithHideMetricPrefixes("debug_", "internal_"),
		WithHideMetricKeys("hidden_key"),
		WithSectionOrder(core.DashboardSectionHeader, core.DashboardSectionToolUsage),
		WithMetricLabels(labels),
		WithCompactLabels(compactLabels),
		WithRawGroups(rawGroup),
		WithSuppressZeroMetricKeys("zero_metric"),
	)

	if dash.ColorRole != core.DashboardColorRoleLavender {
		t.Errorf("ColorRole = %v, want %v", dash.ColorRole, core.DashboardColorRoleLavender)
	}
	if len(dash.GaugePriority) != 2 || dash.GaugePriority[0] != "quota_weekly" {
		t.Errorf("GaugePriority = %v, want [quota_weekly quota_5h]", dash.GaugePriority)
	}
	if dash.GaugeMaxLines != 4 {
		t.Errorf("GaugeMaxLines = %d, want 4", dash.GaugeMaxLines)
	}
	if len(dash.CompactRows) != 1 || dash.CompactRows[0].Label != "Status" {
		t.Errorf("CompactRows = %v, want 1 entry with Label Status", dash.CompactRows)
	}
	if len(dash.HideMetricPrefixes) != 2 || dash.HideMetricPrefixes[0] != "debug_" {
		t.Errorf("HideMetricPrefixes = %v", dash.HideMetricPrefixes)
	}
	if len(dash.HideMetricKeys) != 1 || dash.HideMetricKeys[0] != "hidden_key" {
		t.Errorf("HideMetricKeys = %v", dash.HideMetricKeys)
	}
	if len(dash.StandardSectionOrder) != 2 || dash.StandardSectionOrder[0] != core.DashboardSectionHeader {
		t.Errorf("StandardSectionOrder = %v", dash.StandardSectionOrder)
	}
	if dash.MetricLabelOverrides["cost"] != "Estimated Cost" || dash.MetricLabelOverrides["tokens"] != "Total Tokens" {
		t.Errorf("MetricLabelOverrides = %v", dash.MetricLabelOverrides)
	}
	if dash.CompactMetricLabelOverrides["cost"] != "$" || dash.CompactMetricLabelOverrides["tokens"] != "tok" {
		t.Errorf("CompactMetricLabelOverrides = %v", dash.CompactMetricLabelOverrides)
	}
	if len(dash.RawGroups) == 0 || dash.RawGroups[len(dash.RawGroups)-1].Label != "Raw Dump" {
		t.Errorf("RawGroups = %v, want last entry with Label 'Raw Dump'", dash.RawGroups)
	}
	if len(dash.SuppressZeroMetricKeys) != 1 || dash.SuppressZeroMetricKeys[0] != "zero_metric" {
		t.Errorf("SuppressZeroMetricKeys = %v", dash.SuppressZeroMetricKeys)
	}
}

func TestCodingToolDashboard_DefaultsAndOptions(t *testing.T) {
	// Base CodingToolDashboard
	cd := CodingToolDashboard(
		nil, // Nil safety check
		WithColorRole(core.DashboardColorRoleGreen),
	)

	if !cd.ShowClientComposition {
		t.Error("expected ShowClientComposition = true")
	}
	if cd.ClientCompositionHeading != "Clients" {
		t.Errorf("ClientCompositionHeading = %q, want 'Clients'", cd.ClientCompositionHeading)
	}
	if !cd.ShowActualToolUsage {
		t.Error("expected ShowActualToolUsage = true")
	}
	if !cd.ShowMCPUsage {
		t.Error("expected ShowMCPUsage = true")
	}
	if !cd.ShowLanguageComposition {
		t.Error("expected ShowLanguageComposition = true")
	}
	if !cd.ShowCodeStatsComposition {
		t.Error("expected ShowCodeStatsComposition = true")
	}
	if cd.ColorRole != core.DashboardColorRoleGreen {
		t.Errorf("ColorRole = %v, want Green from option", cd.ColorRole)
	}

	// Verify shared code-stats labels are mapped
	if len(cd.MetricLabelOverrides) == 0 {
		t.Error("MetricLabelOverrides should contain shared CodeStatsMetricLabels")
	}
	if len(cd.CompactMetricLabelOverrides) == 0 {
		t.Error("CompactMetricLabelOverrides should contain shared CodeStatsCompactLabels")
	}
}

func TestBase_Concurrency(t *testing.T) {
	b := New(core.ProviderSpec{
		ID: "concurrent-prov",
		Info: core.ProviderInfo{
			Name:   "Concurrent",
			DocURL: "https://example.com",
		},
		Dashboard: DefaultDashboard(WithColorRole(core.DashboardColorRoleBlue)),
		Detail:    core.DefaultDetailWidget(),
	})

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = b.ID()
			_ = b.Describe()
			_ = b.Spec()
			_ = b.Client()
			_ = b.DashboardWidget()
			_ = b.DetailWidget()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

