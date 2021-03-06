package main

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -L /usr/local/Qredo-Crypto-Library/qredolib/build/gflags/lib/ -lgflags
// #cgo LDFLAGS: -L /usr/local/Qredo-Crypto-Library/qredolib/build/protobuf/lib/ -lprotobuf
// #cgo LDFLAGS: -L /usr/local/Qredo-Crypto-Library/qredolib/build/src/proto/ -lqredo_io
// #cgo LDFLAGS: -L /usr/local/Qredo-Crypto-Library/qredolib/build/src/qredo/ -lqredo
// #cgo LDFLAGS: -L /usr/local/Qredo-Crypto-Library/qredolib/build/src/qredo_api/ -lqredo_api
// #include <stdio.h>
// #include <stdlib.h>
// #include "/usr/local/Qredo-Crypto-Library/qredolib/src/qredo_api/qredo_api.h"
import "C"
import "unsafe"
import "fmt"
import "errors"
import log "github.com/Sirupsen/logrus"


//------------------------------------------

//!
//! \brief The system parameters
//!
type Parameters struct {
	// number of participants
	numberOfParticipants uint
	// ring signature threshold
	threshold uint
}

//------------------------------------------

// InitContext - set up new ring
//! \brief initializes the context, this should happen only once
//!
//! \param parameters input  : the system parameters
//! \param result output     : True if initialized correctly, false otherwise
//!
func InitContext(p Parameters) bool {

	// create parameters struct and fill it in
	parameters_extended := C.qredo_parameters{}

	C.qredo_init_parameters(
		&parameters_extended,
		C.ulong(p.numberOfParticipants),
		C.ulong(p.threshold))

	result := C.qredo_init_context(parameters_extended)

	if result == C.int(1) {
		return true
	}

	return false
}

//------------------------------------------

//Keygen - Creates keys
//! \brief Creates the public key and private key
//!
//! \param public_key output : the public key created
//! \param secret_key output : the private key created
//!
func Keygen() (publicKey []byte, privateKey[]byte) {

	// get length of the parameters
	var public_key_length C.ulong
	var private_key_maxlength C.ulong

	C.qredo_get_public_private_key_sizes(
		&public_key_length,
		&private_key_maxlength)

	ptr_public_key := C.malloc(C.size_t(public_key_length))
	defer C.free(ptr_public_key)

	ptr_private_key := C.malloc(C.size_t(private_key_maxlength))
	defer C.free(ptr_private_key)

	var ptr_private_key_length C.ulong

	// call the C API for creating keygen
	result := C.qredo_ring_keygen(
		(*C.uchar)(ptr_public_key),
		(*C.uchar)(ptr_private_key),
		&ptr_private_key_length)

	if result != 0 {
		log.Warn("Failed to create public private key")
	}

	publicKey = C.GoBytes(
		ptr_public_key,
		C.int(public_key_length))

	privateKey = C.GoBytes(
		ptr_private_key,
		C.int(ptr_private_key_length))

	return publicKey, privateKey
}

//------------------------------------------

//!
//! \brief Takes a message creates signature out of it
//!
//! \param message     input  : the message to be signed,
//! \param private_key input  : the private key of the participant
//! \param signers     input  : array of indices with the signers
//! \param public_keys input  : array of all the public keys concatanated
//! \param signature   output : the result signature buffer
//!
func ParticipantSign(message []byte, private_key []byte, signers []uint, public_keys []byte) []byte {

	// TODO check if signers is set correctly
	signature_length := C.qredo_get_participant_signature_size()
	ptr_signature := C.malloc(C.size_t(signature_length))
	defer C.free(ptr_signature)

	number_of_signers := len(signers)
	c_signers := make([]C.ulong, number_of_signers)

	for i, v := range signers {
		c_signers[i] = C.ulong(v)
	}

	result := C.qredo_ring_participant_sign(
		(*C.uchar)(ptr_signature),
		(*C.uchar)(unsafe.Pointer(&message[0])),
		C.ulong(len(message)),
		(*C.uchar)(unsafe.Pointer(&private_key[0])),
		C.ulong(len(private_key)),
		(*C.ulong)(unsafe.Pointer(&c_signers[0])),
		(*C.uchar)(unsafe.Pointer(&public_keys[0])))

	if result != 0 {
		log.Warn("Failed to participant_sign")
	}

	signature := C.GoBytes(
		ptr_signature,
		C.int(signature_length))

	return signature
}

//------------------------------------------

//!
//! \brief Takes an threshold-concatenated participant signatures and creates a ring leader signature
//!
//! \param message                input  : the message to be signed,
//! \param leader_index           input  : index of the leader participant signer
//! \param private_key            input  : leader participant private key
//! \param signers                input  : array of indices with the signers
//! \param participant_signatures input  : concatenated participants signatures
//! \param public_keys            input  : array of all the public keys concatanated
//! \param signature              output : the ring leader signature
//!
func leader_sign(message []byte, leader_index uint, private_key []byte, signers []uint, public_keys []byte, participant_signatures []byte) (lSig []byte, err error) {

	signature_length := C.qredo_get_ring_signature_size()
	ptr_signature := C.malloc(C.size_t(signature_length))
	defer C.free(ptr_signature)

	number_of_signers := len(signers)
	c_signers := make([]C.ulong, number_of_signers)

	for i, v := range signers {
		c_signers[i] = C.ulong(v)
	}

	result := C.qredo_ring_leader_sign(
		(*C.uchar)(ptr_signature),
		C.ulong(leader_index),
		(*C.uchar)(unsafe.Pointer(&message[0])),
		C.ulong(len(message)),
		(*C.uchar)(unsafe.Pointer(&private_key[0])),
		C.ulong(len(private_key)),
		(*C.ulong)(unsafe.Pointer(&c_signers[0])),
		(*C.uchar)(unsafe.Pointer(&public_keys[0])),
		(*C.uchar)(unsafe.Pointer(&participant_signatures[0])))

	if result != 0 {
		err = errors.New("Failed to leader_sign")
	}

	signature := C.GoBytes(
		ptr_signature,
		C.int(signature_length))

	return signature, err
}

//------------------------------------------

//!
//! \brief Verifies that the message was correctly signed
//!
//! \param message           input  : the message to be signed,
//! \param ring_signature    input  : the ring signature created by the leader sign
//! \param public_keys       input  : array of all the public keys concatanated
//! \param result            output : true if given message was verified, false otherwise
//!
func verify(message []byte, ring_signature []byte, public_keys []byte) bool {

	result := C.qredo_ring_verify(
		(*C.uchar)(unsafe.Pointer(&message[0])),
		C.ulong(len(message)),
		(*C.uchar)(unsafe.Pointer(&ring_signature[0])),
		(*C.uchar)(unsafe.Pointer(&public_keys[0])))

	if result == C.int(1) {
		return true
	}

	return false
}

//------------------------------------------

//TrsTest - test threshold-ring signature scheme
	func TrsTest(){
	// func main(){	
	p := Parameters{numberOfParticipants: 10, threshold: 5}
	InitContext(p)

	// concatanated public keys
	var public_keys []byte

	// msgStr := "A4C044F3977995C2CA3D23CC0117BF0DFC2ACF2301F2CBACBDC001D0AB4D6641"

	message := []byte{10, 0}

	leader := uint(2)

	signers := []uint{0, 1, 2, 3, 4}

	private_keys := make([][]byte, p.numberOfParticipants)

	// create public/private keys
	for i := uint(0); i < p.numberOfParticipants; i++ {
		public_key, private_key := Keygen()
		private_keys[i] = private_key
		// concat public keys together
		public_keys = append(public_keys, public_key...)
		fmt.Printf("%v public_key %v = \n", i, public_key)
		fmt.Println("private_key = ", private_key)
	}

	var participants_signatures []byte

	// participant sign
	for i := uint(0); i < p.threshold; i++ {
		signer_index := signers[i]
		participant_signature := ParticipantSign(message, private_keys[signer_index], signers, public_keys)
		fmt.Println("participant_signature = ", participant_signature)
		participants_signatures = append(participants_signatures, participant_signature...)
	}

	// leader/ring signature
	leader_signature, _ := leader_sign(message, leader, private_keys[leader], signers, public_keys, participants_signatures)

	fmt.Println("leader_signature = ", leader_signature)

	// verify
	result := verify(message, leader_signature, public_keys)

	fmt.Println(result)
}
