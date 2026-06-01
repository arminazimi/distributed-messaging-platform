package message

import (
	"gateway/testutil"
	"testing"

	"gateway/internal/model"
)

func TestUpdateMessage_InsertAndHistory(t *testing.T) {
	ctx := testutil.EnsureSetup(t)
	s := model.Message{CustomerID: 1, Recipients: []string{"+1", "+2"}, Type: model.NORMAL, MessageIdentifier: "id-1"}

	if err := InsertPending(ctx, s); err != nil {
		t.Fatalf("insert pending err: %v", err)
	}

	history, err := GetUserHistory(ctx, "1", string(Pending), "id-1")
	if err != nil {
		t.Fatalf("get history err: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(history))
	}
	for _, h := range history {
		if h.MessageIdentifier != "id-1" || h.Status != Pending {
			t.Fatalf("unexpected history %+v", h)
		}
	}
}

func TestUpdateMessage_NoRecipients(t *testing.T) {
	ctx := testutil.EnsureSetup(t)
	s := model.Message{CustomerID: 1, Recipients: nil, Type: model.NORMAL}
	if err := InsertPending(ctx, s); err == nil {
		t.Fatalf("expected error for no recipients")
	}
}

func TestSendMessage_Success(t *testing.T) {
	ctx := testutil.EnsureSetup(t)

	s := model.Message{CustomerID: 1, Recipients: []string{"+1", "+2"}, Type: model.NORMAL, MessageIdentifier: "succ-1"}
	if err := InsertPending(ctx, s); err != nil {
		t.Fatalf("insert pending err: %v", err)
	}
	if err := sendMessage(ctx, s); err != nil {
		t.Fatalf("sendMessage err: %v", err)
	}

	// After processing, final state should be DONE.
	doneRows, _ := GetUserHistory(ctx, "1", string(Done), "succ-1")
	if len(doneRows) != 2 {
		t.Fatalf("expected 2 done rows, got %d", len(doneRows))
	}
	for _, h := range doneRows {
		if h.Provider != "operatorA" {
			t.Fatalf("unexpected provider %q", h.Provider)
		}
	}
}

func TestSendMessage_NoRecipients(t *testing.T) {
	ctx := testutil.EnsureSetup(t)

	s := model.Message{CustomerID: 3, Recipients: nil, Type: model.NORMAL, MessageIdentifier: "no-recips"}
	if err := sendMessage(ctx, s); err == nil {
		t.Fatalf("expected error for no recipients")
	}
}
