/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package builtin

import (
    // for HBMPC
    "fmt"
    "log"
    "os/exec"

	. "github.com/hyperledger/fabric/core/handlers/endorsement/api"
	. "github.com/hyperledger/fabric/core/handlers/endorsement/api/identities"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

// DefaultEndorsementFactory returns an endorsement plugin factory which returns plugins
// that behave as the default endorsement system chaincode
type DefaultEndorsementFactory struct {
}

// New returns an endorsement plugin that behaves as the default endorsement system chaincode
func (*DefaultEndorsementFactory) New() Plugin {
	return &DefaultEndorsement{}
}

// DefaultEndorsement is an endorsement plugin that behaves as the default endorsement system chaincode
type DefaultEndorsement struct {
	SigningIdentityFetcher
}

// Endorse signs the given payload(ProposalResponsePayload bytes), and optionally mutates it.
// Returns:
// The Endorsement: A signature over the payload, and an identity that is used to verify the signature
// The payload that was given as input (could be modified within this function)
// Or error on failure
func (e *DefaultEndorsement) Endorse(prpBytes []byte, sp *peer.SignedProposal) (*peer.Endorsement, []byte, error) {
	signer, err := e.SigningIdentityForRequest(sp)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed fetching signing identity")
	}
	// serialize the signing identity
	identityBytes, err := signer.Serialize()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not serialize the signing identity")
	}

    // XXX ---- HoneyBadgerMPC ----
    //
    // From https://github.com/initc3/HoneyBadgerMPC/issues/24#issuecomment-417183754:
    //
    // Identify a suitable "snip" point in Hyperledger Fabric endorser code, that lets
    // endorsers communicate with the hbmpc application before approving a transaction.
    // This should be at the endorser, roughly in between when it receives a transaction
    // and when it signs it. Essentially the mpc will act like a "pre-ordering" service
    //
    // TODO Replace "dummy" code below with appropriate code.
    cmd := exec.Command("python3", "-m", "honeybadgermpc.polynomial")
    out, errmsg := cmd.CombinedOutput()
    if errmsg != nil {
        log.Fatalf("cmd.Run() failed with %s\n", errmsg)
    }
    fmt.Printf("combined out:\n%s\n", string(out))
    // -XXX --- HoneyBadgerMPC ----

	// sign the concatenation of the proposal response and the serialized endorser identity with this endorser's key
	signature, err := signer.Sign(append(prpBytes, identityBytes...))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not sign the proposal response payload")
	}
	endorsement := &peer.Endorsement{Signature: signature, Endorser: identityBytes}
	return endorsement, prpBytes, nil
}

// Init injects dependencies into the instance of the Plugin
func (e *DefaultEndorsement) Init(dependencies ...Dependency) error {
	for _, dep := range dependencies {
		sIDFetcher, isSigningIdentityFetcher := dep.(SigningIdentityFetcher)
		if !isSigningIdentityFetcher {
			continue
		}
		e.SigningIdentityFetcher = sIDFetcher
		return nil
	}
	return errors.New("could not find SigningIdentityFetcher in dependencies")
}
