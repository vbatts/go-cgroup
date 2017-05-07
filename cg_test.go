package cgroup

import (
	"os/user"
	"testing"
)

func TestUidGid(t *testing.T) {
	Init()

	var wantTaskUid, wantCtrlUid UID
	var wantTaskGid, wantCtrlGid GID

	curUser, err := user.Current()
	if err == nil {
		wantTaskUid = stringToUID(curUser.Uid)
		wantTaskGid = stringToGID(curUser.Gid)
		wantCtrlUid = stringToUID(curUser.Uid)
		wantCtrlGid = stringToGID(curUser.Gid)
	} else {
		t.Logf("cannot get the current user. fall back to 0.\n")
		wantTaskUid, wantTaskGid, wantCtrlUid, wantCtrlGid = 0, 0, 0, 0
	}

	cg := NewCgroup("test_cgroup")
	if err := cg.SetUIDGID(wantTaskUid, wantTaskGid, wantCtrlUid, wantCtrlGid); err != nil {
		t.Fatalf("cannot set cgroup uids/gids: %v\n", err)
	}

	gotTaskUid, gotTaskGid, gotCtrlUid, gotCtrlGid, err := cg.GetUIDGID()
	if err != nil {
		t.Fatalf("cannot get cgroup uids/gids: %v\n", err)
	}

	if wantTaskUid != gotTaskUid || wantTaskGid != gotTaskGid ||
		wantCtrlUid != gotCtrlUid || wantCtrlGid != gotCtrlGid {
		t.Fatalf("wanted (%d,%d,%d,%d), got (%d,%d,%d,%d)\n",
			wantTaskUid, wantTaskGid, wantCtrlUid, wantCtrlGid,
			gotTaskUid, gotTaskGid, gotCtrlUid, gotCtrlGid)
	}
}
