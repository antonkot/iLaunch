package system

import (
	"context"
	"errors"
	"testing"
)

type fakeCommander struct {
	paths map[string]string
	out   []byte
	err   error
}

func (f fakeCommander) LookPath(file string) (string, error) {
	if p, ok := f.paths[file]; ok {
		return p, nil
	}
	return "", errors.New("missing")
}

func (f fakeCommander) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.out, nil
}

func TestCheckEnvironment(t *testing.T) {
	res, err := CheckEnvironment(context.Background(), fakeCommander{
		paths: map[string]string{"node": "/usr/bin/node", "npm": "/usr/bin/npm"},
		out:   []byte("v18.16.0\n"),
	})
	if err != nil {
		t.Fatalf("CheckEnvironment() error = %v", err)
	}
	if res.PackageMgr != "npm" {
		t.Fatalf("expected npm, got %s", res.PackageMgr)
	}
}

func TestCheckEnvironmentVersionTooLow(t *testing.T) {
	_, err := CheckEnvironment(context.Background(), fakeCommander{
		paths: map[string]string{"node": "/usr/bin/node", "pnpm": "/usr/bin/pnpm"},
		out:   []byte("v16.0.0\n"),
	})
	if err == nil {
		t.Fatal("expected version error")
	}
}
