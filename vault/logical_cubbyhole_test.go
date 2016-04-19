package vault

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/logical"
)

func TestCubbyholeBackend_RootPaths(t *testing.T) {
	b := testCubbyholeBackend()
	root := b.SpecialPaths()
	if root != nil {
		t.Fatalf("unexpected: %v", root)
	}
}

func TestCubbyholeBackend_Write(t *testing.T) {
	b := testCubbyholeBackend()
	req := logical.TestRequest(t, logical.UpdateOperation, "foo")
	clientToken, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	req.ClientToken = clientToken
	storage := req.Storage
	req.Data["raw"] = "test"

	resp, err := b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp != nil {
		t.Fatalf("bad: %v", resp)
	}

	req = logical.TestRequest(t, logical.ReadOperation, "foo")
	req.Storage = storage
	req.ClientToken = clientToken
	_, err = b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
}

func TestCubbyholeBackend_Read(t *testing.T) {
	b := testCubbyholeBackend()
	req := logical.TestRequest(t, logical.UpdateOperation, "foo")
	req.Data["raw"] = "test"
	storage := req.Storage
	clientToken, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	req.ClientToken = clientToken

	if _, err := b.HandleRequest(req); err != nil {
		t.Fatalf("err: %v", err)
	}

	req = logical.TestRequest(t, logical.ReadOperation, "foo")
	req.Storage = storage
	req.ClientToken = clientToken

	resp, err := b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	expected := &logical.Response{
		Data: map[string]interface{}{
			"raw": "test",
		},
	}

	if !reflect.DeepEqual(resp, expected) {
		t.Fatalf("bad response.\n\nexpected: %#v\n\nGot: %#v", expected, resp)
	}
}

func TestCubbyholeBackend_Delete(t *testing.T) {
	b := testCubbyholeBackend()
	req := logical.TestRequest(t, logical.UpdateOperation, "foo")
	req.Data["raw"] = "test"
	storage := req.Storage
	clientToken, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	req.ClientToken = clientToken

	if _, err := b.HandleRequest(req); err != nil {
		t.Fatalf("err: %v", err)
	}

	req = logical.TestRequest(t, logical.DeleteOperation, "foo")
	req.Storage = storage
	req.ClientToken = clientToken
	resp, err := b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp != nil {
		t.Fatalf("bad: %v", resp)
	}

	req = logical.TestRequest(t, logical.ReadOperation, "foo")
	req.Storage = storage
	req.ClientToken = clientToken
	resp, err = b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp != nil {
		t.Fatalf("bad: %v", resp)
	}
}

func TestCubbyholeBackend_CIDRWrite(t *testing.T) {
	b := testCubbyholeBackend()
	req := logical.TestRequest(t, logical.UpdateOperation, "foo")
	req.Data["value"] = "test"
	req.Data["cidr_block"] = "192.168.1.0/24"
	storage := req.Storage
	clientToken, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	req.ClientToken = clientToken

	if _, err := b.HandleRequest(req); err != nil {
		t.Fatalf("err: %v", err)
	}

	req = logical.TestRequest(t, logical.ReadOperation, "foo")
	req.Connection = &logical.Connection{RemoteAddr: "192.168.1.0"}
	req.Storage = storage
	req.ClientToken = clientToken

	resp, err := b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	expected := &logical.Response{
		Data: map[string]interface{}{
			"value": "test",
			"cidr_block": "192.168.1.0/24",
		},
	}

	if !reflect.DeepEqual(resp, expected) {
		t.Fatalf("bad response.\n\nexpected: %#v\n\nGot: %#v", expected, resp)
	}

	req.Connection = &logical.Connection{RemoteAddr: "192.168.2.0"}
	resp, err = b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	expected = &logical.Response{
		Data: map[string]interface{}{
			"error":"your source address is not allowed to read this key",
		},
	}

	if !reflect.DeepEqual(resp, expected) {
		t.Fatalf("bad response.\n\nexpected: %#v\n\nGot: %#v", expected, resp)
	}

}

func TestCubbyholeBackend_InvalidCIDRWrite(t *testing.T) {
	b := testCubbyholeBackend()
	req := logical.TestRequest(t, logical.UpdateOperation, "foo")
	req.Data["value"] = "test"
	req.Data["cidr_block"] = "192.168.1:0/24"
	clientToken, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	req.ClientToken = clientToken

	_, err = b.HandleRequest(req)

	expected := "invalid cidr_block specified: invalid CIDR address: 192.168.1:0/24"
	if err.Error() != expected {
		t.Fatalf("bad response.\n\nexpected: %#v\n\nGot: %#v", expected, err)
	}
}


func TestCubbyholeBackend_List(t *testing.T) {
	b := testCubbyholeBackend()
	req := logical.TestRequest(t, logical.UpdateOperation, "foo")
	clientToken, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	req.Data["raw"] = "test"
	req.ClientToken = clientToken
	storage := req.Storage

	if _, err := b.HandleRequest(req); err != nil {
		t.Fatalf("err: %v", err)
	}

	req = logical.TestRequest(t, logical.UpdateOperation, "bar")
	req.Data["raw"] = "baz"
	req.ClientToken = clientToken
	req.Storage = storage

	if _, err := b.HandleRequest(req); err != nil {
		t.Fatalf("err: %v", err)
	}

	req = logical.TestRequest(t, logical.ListOperation, "")
	req.Storage = storage
	req.ClientToken = clientToken
	resp, err := b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	expKeys := []string{"foo", "bar"}
	respKeys := resp.Data["keys"].([]string)
	sort.Strings(expKeys)
	sort.Strings(respKeys)
	if !reflect.DeepEqual(respKeys, expKeys) {
		t.Fatalf("bad response.\n\nexpected: %#v\n\nGot: %#v", expKeys, respKeys)
	}
}

func TestCubbyholeIsolation(t *testing.T) {
	b := testCubbyholeBackend()

	clientTokenA, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	clientTokenB, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	var storageA logical.Storage
	var storageB logical.Storage

	// Populate and test A entries
	req := logical.TestRequest(t, logical.UpdateOperation, "foo")
	req.ClientToken = clientTokenA
	storageA = req.Storage
	req.Data["raw"] = "test"

	resp, err := b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp != nil {
		t.Fatalf("bad: %v", resp)
	}

	req = logical.TestRequest(t, logical.ReadOperation, "foo")
	req.Storage = storageA
	req.ClientToken = clientTokenA
	resp, err = b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	expected := &logical.Response{
		Data: map[string]interface{}{
			"raw": "test",
		},
	}

	if !reflect.DeepEqual(resp, expected) {
		t.Fatalf("bad response.\n\nexpected: %#v\n\nGot: %#v", expected, resp)
	}

	// Populate and test B entries
	req = logical.TestRequest(t, logical.UpdateOperation, "bar")
	req.ClientToken = clientTokenB
	storageB = req.Storage
	req.Data["raw"] = "baz"

	resp, err = b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp != nil {
		t.Fatalf("bad: %v", resp)
	}

	req = logical.TestRequest(t, logical.ReadOperation, "bar")
	req.Storage = storageB
	req.ClientToken = clientTokenB
	resp, err = b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	expected = &logical.Response{
		Data: map[string]interface{}{
			"raw": "baz",
		},
	}

	if !reflect.DeepEqual(resp, expected) {
		t.Fatalf("bad response.\n\nexpected: %#v\n\nGot: %#v", expected, resp)
	}

	// We shouldn't be able to read A from B and vice versa
	req = logical.TestRequest(t, logical.ReadOperation, "foo")
	req.Storage = storageB
	req.ClientToken = clientTokenB
	resp, err = b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp != nil {
		t.Fatalf("err: was able to read from other user's cubbyhole")
	}

	req = logical.TestRequest(t, logical.ReadOperation, "bar")
	req.Storage = storageA
	req.ClientToken = clientTokenA
	resp, err = b.HandleRequest(req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp != nil {
		t.Fatalf("err: was able to read from other user's cubbyhole")
	}
}

func testCubbyholeBackend() logical.Backend {
	b, _ := CubbyholeBackendFactory(&logical.BackendConfig{
		Logger: nil,
		System: logical.StaticSystemView{
			DefaultLeaseTTLVal: time.Hour * 24,
			MaxLeaseTTLVal:     time.Hour * 24 * 30,
		},
	})
	return b
}
