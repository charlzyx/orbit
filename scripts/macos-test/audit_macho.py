#!/usr/bin/env python3
import json, struct, re, sys
from pathlib import Path

def inspect(path):
 data=Path(path).read_bytes()
 magic,cpu,subtype,kind,ncmds,sizeofcmds,flags,reserved=struct.unpack_from('<8I',data)
 assert magic == 0xfeedfacf, 'Not thin little-endian Mach-O 64'
 result={'architecture':{0x0100000c:'arm64',0x01000007:'amd64'}.get(cpu,hex(cpu)),'file_type':kind,'libraries':[],'minimum_os':None,'has_code_signature_command':False}
 syms=[]; off=32
 version=lambda v:f'{v>>16}.{(v>>8)&255}.{v&255}'
 for _ in range(ncmds):
  cmd,size=struct.unpack_from('<II',data,off)
  if cmd in (0xc,0x80000018,0x8000001f,0x20,0x80000023):
   nameoff=struct.unpack_from('<I',data,off+8)[0]; result['libraries'].append(data[off+nameoff:off+size].split(b'\0')[0].decode())
  elif cmd == 0x24: result['minimum_os']=version(struct.unpack_from('<I',data,off+8)[0])
  elif cmd == 0x32: result['minimum_os']=version(struct.unpack_from('<I',data,off+12)[0])
  elif cmd == 0x1d: result['has_code_signature_command']=True
  elif cmd == 0x2:
   symoff,nsyms,stroff,strsize=struct.unpack_from('<4I',data,off+8)
   for i in range(nsyms):
    strx,typ,sect,desc,value=struct.unpack_from('<IBBHQ',data,symoff+i*16)
    if (typ&0x0e) != 0 and typ&1:
     syms.append(data[stroff+strx:stroff+strsize].split(b'\0')[0].decode())
  off+=size
 assert off==32+sizeofcmds
 result['defined_external_symbol_count']=len(syms)
 return result,syms
