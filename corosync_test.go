package corosync

import (
	"fmt"
	"testing"
)

// Here is an example output from a healthy cluster:
//Quorum information
//------------------
//Date:             Tue Oct 29 14:11:19 2019
//Quorum provider:  corosync_votequorum
//Nodes:            3
//Node ID:          2
//Ring ID:          1/12
//Quorate:          Yes
//
//Votequorum information
//----------------------
//Expected votes:   3
//Highest expected: 3
//Total votes:      3
//Quorum:           2
//Flags:            Quorate
//
//Membership information
//----------------------
//    Nodeid      Votes    Qdevice Name
//         1          1         NR 192.168.123.11
//         2          1         NR 192.168.123.12 (local)
//         3          1         NR 192.168.123.13

func TestGetQuorate(t *testing.T) {
	output := `
Quorum information
------------------
Date:             Tue Oct 29 14:11:19 2019
Quorum provider:  corosync_votequorum
Nodes:            3
Node ID:          2
Ring ID:          1/12
Quorate:          %s

Votequorum information
----------------------
`
	q, err := getQuorate(fmt.Sprintf(output, "Yes"))
	if err != nil {
		t.Errorf("Expected nil error, got %#v", err)
	}
	if !q {
		t.Errorf("Expected cluster to be quorate")
	}

	q, err = getQuorate(fmt.Sprintf(output, "No"))
	if err != nil {
		t.Errorf("Expected nil error, got %#v", err)
	}
	if q {
		t.Errorf("Expected cluster to not be quorate")
	}

	q, err = getQuorate(fmt.Sprintf(output, "Something invalid"))
	if err == nil {
		t.Errorf("Expected to get error, got nil")
	}
}

func TestGetVotes(t *testing.T) {
	output := `
Quorate:          Yes

Votequorum information
----------------------
Expected votes:   12345
Highest expected: 12345
Total votes:      %s
Quorum:           12345
Flags:            Quorate

Membership information
----------------------
`
	v, err := getVotes(fmt.Sprintf(output, "3"))
	if err != nil {
		t.Errorf("Expected nil error, got %#v", err)
	}
	if v != 3 {
		t.Errorf("Expected total votes to be 3, got %d", v)
	}

	v, err = getVotes(fmt.Sprintf(output, "134"))
	if err != nil {
		t.Errorf("Expected nil error, got %#v", err)
	}
	if v != 134 {
		t.Errorf("Expected total votes to be 134, got %d", v)
	}

	v, err = getVotes(fmt.Sprintf(output, "invalid"))
	if err == nil {
		t.Errorf("Expected to get error, got nil")
	}
}

func TestGetVotesExpected(t *testing.T) {
	output := `
Quorate:          Yes

Votequorum information
----------------------
Expected votes:   %s
Highest expected: 12345
Total votes:      12345
Quorum:           12345
Flags:            Quorate

Membership information
----------------------
`
	v, err := getVotesExpected(fmt.Sprintf(output, "3"))
	if err != nil {
		t.Errorf("Expected nil error, got %#v", err)
	}
	if v != 3 {
		t.Errorf("Expected total votes to be 3, got %d", v)
	}

	v, err = getVotesExpected(fmt.Sprintf(output, "134"))
	if err != nil {
		t.Errorf("Expected nil error, got %#v", err)
	}
	if v != 134 {
		t.Errorf("Expected total votes to be 134, got %d", v)
	}

	v, err = getVotesExpected(fmt.Sprintf(output, "invalid"))
	if err == nil {
		t.Errorf("Expected to get error, got nil")
	}
}
