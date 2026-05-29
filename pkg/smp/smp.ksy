meta:
  id: smp
  endian: be
  bit-endian: be
types:
  message:
    seq:
      - id: reserved
        type: b3
      - id: ver
        type: b2
      - id: op
        type: b3
        enum: operation
      - id: flags
        type: u1
      - id: length
        type: u2
      - id: group
        type: u2
        enum: group
      - id: seq
        type: u1
      - id: body
        type:
          switch-on: group
          cases:
            'group::os': os
            'group::image': image
            'group::stat': stat
            'group::settings': settings
            'group::log': generic
            'group::crash': generic
            'group::split': generic
            'group::run': generic
            'group::fs': fs
            'group::shell': shell
            'group::zephyr': zephyr
  os:
    seq:
      - id: command
        type: u1
        enum: os_command
      - id: payload
        size: _parent.length
  image:
    seq:
      - id: command
        type: u1
        enum: image_command
      - id: payload
        size: _parent.length
  stat:
    seq:
      - id: command
        type: u1
        enum: stat_command
      - id: payload
        size: _parent.length
  settings:
    seq:
      - id: command
        type: u1
        enum: settings_command
      - id: payload
        size: _parent.length
  fs:
    seq:
      - id: command
        type: u1
        enum: fs_command
      - id: payload
        size: _parent.length
  shell:
    seq:
      - id: command
        type: u1
        enum: shell_command
      - id: payload
        size: _parent.length
  zephyr:
    seq:
      - id: command
        type: u1
        enum: zephyr_command
      - id: payload
        size: _parent.length
  generic:
    seq:
      - id: command
        type: u1
      - id: payload
        size: _parent.length
enums:
  operation:
    0: read
    1: read_response
    2: write
    3: write_response
  group:
    0: os
    1: image
    2: stat
    3: settings
    4: log
    5: crash
    6: split
    7: run
    8: fs
    9: shell
    63: zephyr
  os_command:
    0: echo
    1: echo_control
    2: task_stats
    3: mempool_stats
    4: datetime
    5: reset
    6: parameters
    7: os_app_info
    8: bootloader_info
  image_command:
    0: state
    1: upload
    2: file_image
    3: core_list
    4: core_load
    5: erase
    6: slot_info
  stat_command:
    0: group_data
    1: list_groups
  settings_command:
    0: read_write
    1: delete
    2: commit
    3: load_save
  fs_command:
    0: download_upload
    1: status
    2: hash
    3: supported_hash_types
    4: close
  shell_command:
    0: execute
  zephyr_command:
    0: erase

  errors:
    0: ok
    1: unknown
    2: no_mem
    3: inval
    4: timeout
    5: no_ent
    6: bad_state
    7: msg_size
    8: not_sup
    9: corrupt
    10: busy
    11: access_denied
    12: unsupported_too_old
    13: unsupported_too_new
