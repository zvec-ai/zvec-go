//go:build purego || !cgo

package zvec

import (
	"errors"
	"testing"
)

func TestPuregoAPIRetriesAfterLoadFailure(t *testing.T) {
	puregoLoadMu.Lock()
	savedLoaded := puregoLoaded.Load()
	savedHandle := puregoHandle
	savedFns := puregoFns
	savedLoader := puregoBackendLoader
	puregoLoaded.Store(false)
	puregoHandle = 0
	puregoFns = zvecPuregoAPI{}
	attempts := 0
	puregoBackendLoader = func() error {
		attempts++
		if attempts == 1 {
			return errors.New("temporary load failure")
		}
		puregoHandle = 1
		return nil
	}
	puregoLoadMu.Unlock()

	t.Cleanup(func() {
		puregoLoadMu.Lock()
		puregoLoaded.Store(savedLoaded)
		puregoHandle = savedHandle
		puregoFns = savedFns
		puregoBackendLoader = savedLoader
		puregoLoadMu.Unlock()
	})

	if _, err := puregoAPI(); err == nil {
		t.Fatal("first puregoAPI() call returned nil error")
	}
	if _, err := puregoAPI(); err != nil {
		t.Fatalf("second puregoAPI() call failed: %v", err)
	}
	if _, err := puregoAPI(); err != nil {
		t.Fatalf("cached puregoAPI() call failed: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("loader attempts = %d, want 2", attempts)
	}
}

func TestPuregoAPIFastPath(t *testing.T) {
	puregoLoadMu.Lock()
	savedLoaded := puregoLoaded.Load()
	savedHandle := puregoHandle
	savedFns := puregoFns
	savedLoader := puregoBackendLoader
	puregoLoaded.Store(true)
	puregoHandle = 1
	puregoBackendLoader = func() error {
		t.Fatal("loaded fast path called backend loader")
		return nil
	}
	puregoLoadMu.Unlock()

	t.Cleanup(func() {
		puregoLoadMu.Lock()
		puregoLoaded.Store(savedLoaded)
		puregoHandle = savedHandle
		puregoFns = savedFns
		puregoBackendLoader = savedLoader
		puregoLoadMu.Unlock()
	})

	api, err := puregoAPI()
	if err != nil {
		t.Fatalf("puregoAPI() failed: %v", err)
	}
	if api != &puregoFns {
		t.Fatal("puregoAPI() returned an unexpected function table")
	}
}

func TestValidatePuregoVersion(t *testing.T) {
	tests := []struct {
		name                string
		major, minor, patch int32
		wantErr             bool
	}{
		{name: "minimum", major: 0, minor: 5, patch: 1},
		{name: "newer patch", major: 0, minor: 5, patch: 9},
		{name: "older patch", major: 0, minor: 5, patch: 0, wantErr: true},
		{name: "different pre-v1 minor", major: 0, minor: 6, patch: 0, wantErr: true},
		{name: "different major", major: 1, minor: 0, patch: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &zvecPuregoAPI{
				getVersionMajor: func() int32 { return tt.major },
				getVersionMinor: func() int32 { return tt.minor },
				getVersionPatch: func() int32 { return tt.patch },
				checkVersion: func(major, minor, patch int32) bool {
					if tt.major != major {
						return tt.major > major
					}
					if tt.minor != minor {
						return tt.minor > minor
					}
					return tt.patch >= patch
				},
			}
			err := validatePuregoVersion(api)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validatePuregoVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
