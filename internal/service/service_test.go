package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cedar2025/xboard-node/internal/cert"
	"github.com/cedar2025/xboard-node/internal/config"
	"github.com/cedar2025/xboard-node/internal/kernel"
	"github.com/cedar2025/xboard-node/internal/limiter"
	"github.com/cedar2025/xboard-node/internal/model"
	"golang.org/x/time/rate"
)

type fakeKernel struct {
	running bool

	startErr  error
	reloadErr error
	updateErr error
	addErr    error

	startCalls  int
	reloadCalls int
	updateCalls int
	addCalls    int
	removeCalls int
	startUsers  []model.UserSpec

	onUpdateUsers func([]model.UserSpec)
	onAddUsers    func([]model.UserSpec)
	onRemoveUsers func([]model.UserSpec)

	speedLimitFunc  func(string) *rate.Limiter
	deviceLimitFunc func(string) (int, bool)
}

func (f *fakeKernel) Name() string                      { return "fake" }
func (f *fakeKernel) Protocols() []string               { return []string{"vless"} }
func (f *fakeKernel) Capabilities() kernel.Capabilities { return kernel.Capabilities{} }
func (f *fakeKernel) Start(nodeConfig *model.NodeSpec, users []model.UserSpec, tls kernel.TLSCert) error {
	_, _, _ = nodeConfig, users, tls
	f.startCalls++
	if f.startErr != nil {
		return f.startErr
	}
	f.startUsers = append([]model.UserSpec(nil), users...)
	f.running = true
	return nil
}
func (f *fakeKernel) Stop()           { f.running = false }
func (f *fakeKernel) IsRunning() bool { return f.running }
func (f *fakeKernel) Reload(nodeConfig *model.NodeSpec, users []model.UserSpec, tls kernel.TLSCert) error {
	_, _, _ = nodeConfig, users, tls
	f.reloadCalls++
	return f.reloadErr
}

func (f *fakeKernel) AddUsers(users []model.UserSpec) (int, error) {
	f.addCalls++
	if f.onAddUsers != nil {
		f.onAddUsers(users)
	}
	if f.addErr != nil {
		return 0, f.addErr
	}
	return len(users), nil
}
func (f *fakeKernel) RemoveUsers(users []model.UserSpec) (int, error) {
	f.removeCalls++
	if f.onRemoveUsers != nil {
		f.onRemoveUsers(users)
	}
	return len(users), nil
}
func (f *fakeKernel) UpdateUsers(users []model.UserSpec) (int, int, error) {
	f.updateCalls++
	if f.onUpdateUsers != nil {
		f.onUpdateUsers(users)
	}
	if f.updateErr != nil {
		return 0, 0, f.updateErr
	}
	return len(users), 0, nil
}
func (f *fakeKernel) GetUserTraffic(ctx context.Context) (map[int][2]int64, map[int]map[string]bool, int, error) {
	_ = ctx
	return nil, nil, 0, nil
}
func (f *fakeKernel) CloseConnection(ctx context.Context, connID string) error {
	_, _ = ctx, connID
	return nil
}
func (f *fakeKernel) CloseUserConnections(ctx context.Context, uuid string) error {
	_, _ = ctx, uuid
	return nil
}
func (f *fakeKernel) SetSpeedLimitFunc(fn func(uuid string) *rate.Limiter) { f.speedLimitFunc = fn }
func (f *fakeKernel) SetDeviceLimitFunc(fn func(uuid string) (int, bool))  { f.deviceLimitFunc = fn }
func (f *fakeKernel) UpdateGlobalDevices(users map[int][]string)           { _ = users }
func (f *fakeKernel) ClearGlobalDevices()                                  {}

func newTestService(k *fakeKernel) *Service {
	sharedLimiter := limiter.New()
	s := &Service{
		kernel:       k,
		limiter:      sharedLimiter,
		speedTracker: limiter.NewSpeedTracker(sharedLimiter),
		cert:         cert.NewManager(config.CertConfig{}),
	}
	k.SetSpeedLimitFunc(s.speedTracker.GetLimiter)
	k.SetDeviceLimitFunc(s.limiter.GetDeviceLimitByUUID)
	return s
}

func TestApplyConfigUpdateRestoresHashAfterReloadAndRestartFailure(t *testing.T) {
	k := &fakeKernel{
		running:   true,
		reloadErr: errors.New("reload failed"),
		startErr:  errors.New("restart failed"),
	}
	s := newTestService(k)
	s.cfg = &config.Config{}
	s.lastUsers = []model.UserSpec{{ID: 1, UUID: "uuid-1"}}
	oldConfig := &model.NodeSpec{Protocol: "vless", ServerPort: 1001}
	newConfig := &model.NodeSpec{Protocol: "vless", ServerPort: 1002}
	s.lastConfig = oldConfig
	s.lastConfigHash = computeConfigHash(oldConfig)

	if s.applyConfigUpdate(context.Background(), newConfig, computeConfigHash(newConfig)) {
		t.Fatal("expected config apply to fail")
	}
	if s.lastConfig != oldConfig {
		t.Fatalf("lastConfig = %#v, want old config", s.lastConfig)
	}
	if s.lastConfigHash != computeConfigHash(oldConfig) {
		t.Fatalf("lastConfigHash = %q, want old hash", s.lastConfigHash)
	}

	k.reloadErr = nil
	k.startErr = nil
	if !s.applyConfigUpdate(context.Background(), newConfig, computeConfigHash(newConfig)) {
		t.Fatal("expected same config to be retried successfully")
	}
	if s.lastConfigHash != computeConfigHash(newConfig) {
		t.Fatalf("lastConfigHash = %q, want new hash", s.lastConfigHash)
	}
}

func TestTrackAndEnforceRestartsUnexpectedlyStoppedKernel(t *testing.T) {
	k := &fakeKernel{}
	s := newTestService(k)
	s.lastConfig = &model.NodeSpec{Protocol: "vless", ServerPort: 1001}
	s.lastUsers = []model.UserSpec{{ID: 1, UUID: "uuid-1"}}

	s.trackAndEnforce(context.Background())

	if !k.running || k.startCalls != 1 {
		t.Fatalf("kernel running=%v startCalls=%d, want one recovery start", k.running, k.startCalls)
	}
}

func TestApplyUserUpdatePreparesLimiterBeforeKernelUpdate(t *testing.T) {
	k := &fakeKernel{running: true}
	s := newTestService(k)
	s.lastConfig = &model.NodeSpec{Protocol: "vless"}
	oldUsers := []model.UserSpec{{ID: 1, UUID: "uuid-old", SpeedLimit: 4}}
	s.updateUserState(oldUsers)

	newUsers := []model.UserSpec{{ID: 2, UUID: "uuid-new", SpeedLimit: 8}}
	k.onUpdateUsers = func(users []model.UserSpec) {
		if len(users) != 1 || users[0].UUID != "uuid-new" {
			t.Fatalf("unexpected users passed to UpdateUsers: %#v", users)
		}
		if got := k.speedLimitFunc("uuid-new"); got == nil {
			t.Fatal("expected new user's limiter to be visible before kernel UpdateUsers")
		}
	}

	s.applyUserUpdate(context.Background(), newUsers, computeUserHash(newUsers))

	if got := k.updateCalls; got != 1 {
		t.Fatalf("UpdateUsers call count = %d, want 1", got)
	}
	if len(s.lastUsers) != 1 || s.lastUsers[0].UUID != "uuid-new" {
		t.Fatalf("lastUsers = %#v, want new users", s.lastUsers)
	}
	if s.speedTracker.GetLimiter("uuid-new") == nil {
		t.Fatal("expected limiter for new user after successful update")
	}
}

func TestApplyUserUpdateStartsStoppedKernelForFirstUsers(t *testing.T) {
	k := &fakeKernel{}
	s := newTestService(k)
	s.lastConfig = &model.NodeSpec{Protocol: "vless"}

	users := []model.UserSpec{{ID: 1, UUID: "uuid-new", SpeedLimit: 8}}
	s.applyUserUpdate(context.Background(), users, computeUserHash(users))

	if got := k.startCalls; got != 1 {
		t.Fatalf("Start call count = %d, want 1", got)
	}
	if got := k.updateCalls; got != 0 {
		t.Fatalf("UpdateUsers call count = %d, want 0 after starting with complete user set", got)
	}
	if len(k.startUsers) != 1 || k.startUsers[0].UUID != "uuid-new" {
		t.Fatalf("started users = %#v, want new user", k.startUsers)
	}
	if !k.running {
		t.Fatal("kernel is not running after first user update")
	}
	if len(s.lastUsers) != 1 || s.lastUsers[0].UUID != "uuid-new" {
		t.Fatalf("lastUsers = %#v, want new users", s.lastUsers)
	}
}

func TestApplyUserUpdateRetriesAfterInitialKernelStartFailure(t *testing.T) {
	k := &fakeKernel{startErr: errors.New("start failed")}
	s := newTestService(k)
	s.lastConfig = &model.NodeSpec{Protocol: "vless"}
	oldUsers := []model.UserSpec{{ID: 1, UUID: "uuid-old", SpeedLimit: 4}}
	s.updateUserState(oldUsers)
	oldHash := s.lastUserHash

	newUsers := []model.UserSpec{{ID: 2, UUID: "uuid-new", SpeedLimit: 8}}
	s.applyUserUpdate(context.Background(), newUsers, computeUserHash(newUsers))

	if got := k.startCalls; got != 1 {
		t.Fatalf("Start call count = %d, want 1", got)
	}
	if len(s.lastUsers) != 1 || s.lastUsers[0].UUID != "uuid-old" {
		t.Fatalf("lastUsers = %#v, want old users after failed start", s.lastUsers)
	}
	if s.lastUserHash != oldHash {
		t.Fatalf("lastUserHash = %q, want %q for retry", s.lastUserHash, oldHash)
	}
	if s.speedTracker.GetLimiter("uuid-new") != nil {
		t.Fatal("expected failed user's limiter to be removed after rollback")
	}
}

func TestApplyUserDeltaStartsStoppedKernelForFirstUser(t *testing.T) {
	k := &fakeKernel{}
	s := newTestService(k)
	s.lastConfig = &model.NodeSpec{Protocol: "vless"}

	delta := []model.UserSpec{{ID: 1, UUID: "uuid-new", SpeedLimit: 8}}
	s.applyUserDelta(context.Background(), "add", delta)

	if got := k.startCalls; got != 1 {
		t.Fatalf("Start call count = %d, want 1", got)
	}
	if got := k.addCalls; got != 0 {
		t.Fatalf("AddUsers call count = %d, want 0 after starting with complete user set", got)
	}
	if len(k.startUsers) != 1 || k.startUsers[0].UUID != "uuid-new" {
		t.Fatalf("started users = %#v, want new user", k.startUsers)
	}
	if !k.running {
		t.Fatal("kernel is not running after first user delta")
	}
}

func TestApplyUserDeltaRetriesAfterInitialKernelStartFailure(t *testing.T) {
	k := &fakeKernel{startErr: errors.New("start failed")}
	s := newTestService(k)
	s.lastConfig = &model.NodeSpec{Protocol: "vless"}
	oldUsers := []model.UserSpec{{ID: 1, UUID: "uuid-old", SpeedLimit: 4}}
	s.updateUserState(oldUsers)
	oldHash := s.lastUserHash

	s.applyUserDelta(
		context.Background(),
		"add",
		[]model.UserSpec{{ID: 2, UUID: "uuid-new", SpeedLimit: 8}},
	)

	if got := k.startCalls; got != 1 {
		t.Fatalf("Start call count = %d, want 1", got)
	}
	if len(s.lastUsers) != 1 || s.lastUsers[0].UUID != "uuid-old" {
		t.Fatalf("lastUsers = %#v, want old users after failed start", s.lastUsers)
	}
	if s.lastUserHash != oldHash {
		t.Fatalf("lastUserHash = %q, want %q for retry", s.lastUserHash, oldHash)
	}
	if s.speedTracker.GetLimiter("uuid-new") != nil {
		t.Fatal("expected failed delta user's limiter to be removed after rollback")
	}
}

func TestApplyUserUpdateRestoresStateWhenKernelAndRestartFail(t *testing.T) {
	k := &fakeKernel{
		running:   true,
		updateErr: errors.New("update failed"),
		startErr:  errors.New("restart failed"),
	}
	s := newTestService(k)
	s.lastConfig = &model.NodeSpec{Protocol: "vless"}
	oldUsers := []model.UserSpec{{ID: 1, UUID: "uuid-old", SpeedLimit: 4}}
	s.updateUserState(oldUsers)
	oldHash := s.lastUserHash

	newUsers := []model.UserSpec{{ID: 2, UUID: "uuid-new", SpeedLimit: 8}}
	s.applyUserUpdate(context.Background(), newUsers, computeUserHash(newUsers))

	if got := k.startCalls; got != 1 {
		t.Fatalf("Start call count = %d, want 1", got)
	}
	if len(s.lastUsers) != 1 || s.lastUsers[0].UUID != "uuid-old" {
		t.Fatalf("lastUsers = %#v, want restored old users", s.lastUsers)
	}
	if s.lastUserHash != oldHash {
		t.Fatalf("lastUserHash = %q, want %q", s.lastUserHash, oldHash)
	}
	if s.speedTracker.GetLimiter("uuid-old") == nil {
		t.Fatal("expected old limiter to be restored after rollback")
	}
	if s.speedTracker.GetLimiter("uuid-new") != nil {
		t.Fatal("expected new limiter to be removed after rollback")
	}
}

func TestApplyUserDeltaAddPreparesLimiterBeforeKernelUpdate(t *testing.T) {
	k := &fakeKernel{running: true}
	s := newTestService(k)
	s.lastConfig = &model.NodeSpec{Protocol: "vless"}
	oldUsers := []model.UserSpec{{ID: 1, UUID: "uuid-old", SpeedLimit: 4}}
	s.updateUserState(oldUsers)

	delta := []model.UserSpec{{ID: 2, UUID: "uuid-new", SpeedLimit: 8}}
	k.onAddUsers = func(users []model.UserSpec) {
		if len(users) != 1 || users[0].UUID != "uuid-new" {
			t.Fatalf("unexpected users passed to AddUsers: %#v", users)
		}
		if got := k.speedLimitFunc("uuid-new"); got == nil {
			t.Fatal("expected delta user's limiter to be visible before kernel AddUsers")
		}
	}

	s.applyUserDelta(context.Background(), "add", delta)

	if got := k.addCalls; got != 1 {
		t.Fatalf("AddUsers call count = %d, want 1", got)
	}
	if s.speedTracker.GetLimiter("uuid-new") == nil {
		t.Fatal("expected limiter for delta-added user after successful update")
	}
}

func TestApplyUserDeltaUUIDChangeUsesFullUpdate(t *testing.T) {
	k := &fakeKernel{running: true}
	s := newTestService(k)
	s.lastConfig = &model.NodeSpec{Protocol: "vless"}
	oldUsers := []model.UserSpec{{ID: 15, UUID: "uuid-old", SpeedLimit: 4}}
	s.updateUserState(oldUsers)

	delta := []model.UserSpec{{ID: 15, UUID: "uuid-new", SpeedLimit: 4}}
	k.onUpdateUsers = func(users []model.UserSpec) {
		if len(users) != 1 || users[0].ID != 15 || users[0].UUID != "uuid-new" {
			t.Fatalf("unexpected users passed to UpdateUsers: %#v", users)
		}
	}

	s.applyUserDelta(context.Background(), "add", delta)

	if got := k.updateCalls; got != 1 {
		t.Fatalf("UpdateUsers call count = %d, want 1", got)
	}
	if k.addCalls != 0 || k.removeCalls != 0 {
		t.Fatalf("UUID replacement used incremental calls: add=%d remove=%d", k.addCalls, k.removeCalls)
	}
	if len(s.lastUsers) != 1 || s.lastUsers[0].UUID != "uuid-new" {
		t.Fatalf("lastUsers = %#v, want rotated UUID", s.lastUsers)
	}
}

func TestApplyUserDeltaUUIDChangeRestartsKernelWhenFullUpdateFails(t *testing.T) {
	k := &fakeKernel{
		running:   true,
		updateErr: errors.New("update failed"),
	}
	s := newTestService(k)
	s.lastConfig = &model.NodeSpec{Protocol: "vless"}
	s.updateUserState([]model.UserSpec{{ID: 15, UUID: "uuid-old", SpeedLimit: 4}})

	s.applyUserDelta(
		context.Background(),
		"add",
		[]model.UserSpec{{ID: 15, UUID: "uuid-new", SpeedLimit: 4}},
	)

	if got := k.startCalls; got != 1 {
		t.Fatalf("Start call count = %d, want 1", got)
	}
	if len(k.startUsers) != 1 || k.startUsers[0].UUID != "uuid-new" {
		t.Fatalf("restart users = %#v, want rotated UUID", k.startUsers)
	}
	if len(s.lastUsers) != 1 || s.lastUsers[0].UUID != "uuid-new" {
		t.Fatalf("lastUsers = %#v, want rotated UUID after successful restart", s.lastUsers)
	}
}

func TestApplyUserDeltaUUIDChangeRestoresStateWhenRestartFails(t *testing.T) {
	k := &fakeKernel{
		running:   true,
		updateErr: errors.New("update failed"),
		startErr:  errors.New("restart failed"),
	}
	s := newTestService(k)
	s.lastConfig = &model.NodeSpec{Protocol: "vless"}
	oldUsers := []model.UserSpec{{ID: 15, UUID: "uuid-old", SpeedLimit: 4}}
	s.updateUserState(oldUsers)
	oldHash := s.lastUserHash

	s.applyUserDelta(
		context.Background(),
		"add",
		[]model.UserSpec{{ID: 15, UUID: "uuid-new", SpeedLimit: 4}},
	)

	if got := k.startCalls; got != 1 {
		t.Fatalf("Start call count = %d, want 1", got)
	}
	if len(s.lastUsers) != 1 || s.lastUsers[0].UUID != "uuid-old" {
		t.Fatalf("lastUsers = %#v, want old UUID after rollback", s.lastUsers)
	}
	if s.lastUserHash != oldHash {
		t.Fatalf("lastUserHash = %q, want %q after rollback", s.lastUserHash, oldHash)
	}
	if s.speedTracker.GetLimiter("uuid-new") != nil {
		t.Fatal("expected failed UUID limiter to be removed after rollback")
	}
}

func TestApplyUserDeltaRestartsKernelWhenAddAndFullUpdateFail(t *testing.T) {
	k := &fakeKernel{
		running:   true,
		addErr:    errors.New("add failed"),
		updateErr: errors.New("update failed"),
	}
	s := newTestService(k)
	s.lastConfig = &model.NodeSpec{Protocol: "vless"}
	oldUsers := []model.UserSpec{{ID: 1, UUID: "uuid-old", SpeedLimit: 4}}
	s.updateUserState(oldUsers)

	delta := []model.UserSpec{{ID: 2, UUID: "uuid-new", SpeedLimit: 8}}
	s.applyUserDelta(context.Background(), "add", delta)

	if got := k.startCalls; got != 1 {
		t.Fatalf("Start call count = %d, want 1", got)
	}
	if len(k.startUsers) != 2 {
		t.Fatalf("restart users = %#v, want complete merged user set", k.startUsers)
	}
	if len(s.lastUsers) != 2 || s.speedTracker.GetLimiter("uuid-new") == nil {
		t.Fatalf("service state was not retained after successful restart: %#v", s.lastUsers)
	}
}

func TestValidateNodeRuntimeRejectsUnsupportedDNSProvider(t *testing.T) {
	cfg := &config.Config{Kernel: config.KernelConfig{Type: "singbox"}}
	err := validateNodeRuntime(cfg, []string{"http"}, &model.NodeSpec{
		Protocol: "http",
		CertConfig: &config.CertConfig{
			CertMode:    "dns",
			DNSProvider: "3123123",
			Domain:      "example.com",
		},
	}, kernel.TLSCert{CertPEM: []byte("CERT"), KeyPEM: []byte("KEY")})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() == "" {
		t.Fatal("expected non-empty error")
	}
	if got := err.Error(); !strings.HasPrefix(got, `unsupported cert_config.dns_provider "3123123" (supported: `) {
		t.Fatalf("unexpected error: %v", got)
	}
}

func TestValidateNodeRuntimeAllowsSelfManagedTLSBeforeFilesExist(t *testing.T) {
	cfg := &config.Config{Kernel: config.KernelConfig{Type: "singbox"}}
	err := validateNodeRuntime(cfg, []string{"anytls", "hysteria"}, &model.NodeSpec{
		Protocol: "anytls",
		CertConfig: &config.CertConfig{
			CertMode: "self",
			Domain:   "example.com",
		},
	}, kernel.TLSCert{})
	if err != nil {
		t.Fatalf("expected self-managed TLS config to pass validation, got %v", err)
	}
}

func TestValidateNodeRuntimeAllowsSingboxRealityWithRequiredFields(t *testing.T) {
	cfg := &config.Config{Kernel: config.KernelConfig{Type: "singbox"}}
	err := validateNodeRuntime(cfg, []string{"vless"}, &model.NodeSpec{
		Protocol: "vless",
		TLS:      2,
		TLSSettings: map[string]any{
			"private_key": "test-key",
			"server_name": "example.com",
		},
	}, kernel.TLSCert{CertPEM: []byte("CERT"), KeyPEM: []byte("KEY")})
	if err != nil {
		t.Fatalf("expected sing-box reality validation to pass, got %v", err)
	}
}

func TestValidateNodeRuntimeRejectsRealityWithoutTLSSettings(t *testing.T) {
	cfg := &config.Config{Kernel: config.KernelConfig{Type: "singbox"}}
	err := validateNodeRuntime(cfg, []string{"vless"}, &model.NodeSpec{
		Protocol: "vless",
		TLS:      2,
	}, kernel.TLSCert{CertPEM: []byte("CERT"), KeyPEM: []byte("KEY")})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "reality tls requires tls_settings" {
		t.Fatalf("unexpected error: %v", got)
	}
}

func TestValidateNodeRuntimeRejectsRealityWithoutPrivateKey(t *testing.T) {
	cfg := &config.Config{Kernel: config.KernelConfig{Type: "singbox"}}
	err := validateNodeRuntime(cfg, []string{"vless"}, &model.NodeSpec{
		Protocol: "vless",
		TLS:      2,
		TLSSettings: map[string]any{
			"server_name": "example.com",
		},
	}, kernel.TLSCert{CertPEM: []byte("CERT"), KeyPEM: []byte("KEY")})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "reality tls requires tls_settings.private_key" {
		t.Fatalf("unexpected error: %v", got)
	}
}

func TestValidateNodeRuntimeRejectsRealityWithoutServerNameOrDest(t *testing.T) {
	cfg := &config.Config{Kernel: config.KernelConfig{Type: "singbox"}}
	err := validateNodeRuntime(cfg, []string{"vless"}, &model.NodeSpec{
		Protocol: "vless",
		TLS:      2,
		TLSSettings: map[string]any{
			"private_key": "test-key",
		},
	}, kernel.TLSCert{CertPEM: []byte("CERT"), KeyPEM: []byte("KEY")})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "reality tls requires tls_settings.server_name or tls_settings.dest" {
		t.Fatalf("unexpected error: %v", got)
	}
}
