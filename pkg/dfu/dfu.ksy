meta:
  id: dfu
  endian: le
types:
  package:
    seq:
      - id: header
        type: header
        size: 512
      - id: body
        size: header.len_body
      - id: tlvs
        type: tlvs
  header:
    seq:
      - id: magic
        type: u4
        valid: "2532554813" #0x96f3b83d
      - id: loadaddr
        type: u4
      - id: len_header
        type: u2
      - id: len_protect
        type: u2
      - id: len_body
        type: u4
      - id: flags
        type: u4
      - id: version
        type: version
      - id: pad
        type: u4
  tlvs:
    seq:
      - id: magic
        type: u2
        valid: "26887" #0x6907
      - id: size
        type: u2
      - id: tlv_container
        size: size - 4
        type: tlv_container
  tlv_container:
    seq:
      - id: tlv
        type: tlv
        repeat: eos
  tlv:
    seq:
      - id: type
        type: u1
        enum: tlv_types
      - id: pad
        type: u1
      - id: len_tlv_inner
        type: u2
      - id: tlv_inner
        size: len_tlv_inner
  version:
    seq:
      - id: major
        type: u1
      - id: minor
        type: u1
      - id: revision
        type: u2
      - id: buildnum
        type: u4
enums:
  tlv_types:
    0x01: keyhash
    0x02: pubkey
    0x10: sha256
    0x11: sha384
    0x12: sha512
    0x20: rsa2048
    0x21: ecdsa224
    0x22: ecdsa_sig
    0x23: rsa3072
    0x24: ed25519
    0x25: pure_sig
    0x30: enc_rsa2048
    0x31: enc_kw
    0x32: enc_x25519
    0x33: enc_x25519_sha
    0x40: dependency
    0x50: security_count
    0x60: boot_record
    0x70: decomp_size
    0x71: decomp_sha
    0x72: decomp_sig
    0x73: comp_dec_size
    0x80: vid
    0x81: cid
