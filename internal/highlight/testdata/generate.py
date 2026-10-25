# Regenerates behavioral output fixtures by executing the pinned Rust source.
# No Go implementation code or expectations are read. Requires rustc >= 1.94.
import re, json, hashlib, subprocess, pathlib, os, tempfile
root=pathlib.Path(__file__).resolve().parents[3];src=(root/'crates/orbit-pi/src/highlight.rs').read_text()
assert hashlib.sha256(src.encode()).hexdigest() == 'd4c97a77caa49da5d628be3d7657aa523865c3fc65a43a539859b09e4200e65a', 'Oracle source no longer matches pinned baseline'


fixture_tmp=tempfile.TemporaryDirectory()
fixture_path=pathlib.Path(fixture_tmp.name)
oracle_source='#![allow(dead_code)]\n#[path=ORACLE_SOURCE] mod highlight;\nuse std::io::{self,BufRead};\nfn main(){\n for input in io::stdin().lock().lines(){\n  let input=input.unwrap();let p:Vec<_>=input.split(\'\\t\').collect();\n  let lang=highlight::lang_for_tag(p[0]).unwrap();\n  let carry=match p[1]{"1"=>highlight::Carry::BlockComment,"2"=>highlight::Carry::String(p[2].parse().unwrap()),_=>highlight::Carry::None};\n  let bytes:Vec<u8>=p[3].as_bytes().chunks(2).map(|c|u8::from_str_radix(std::str::from_utf8(c).unwrap(),16).unwrap()).collect();\n  let text=String::from_utf8(bytes).unwrap();let(tokens,next)=highlight::tokenize_line(lang,&text,carry);\n  match next{highlight::Carry::None=>print!("0\\t0"),highlight::Carry::BlockComment=>print!("1\\t0"),highlight::Carry::String(n)=>print!("2\\t{}",n)};\n  for t in tokens{print!("\\t{},{},{}",t.range.start,t.range.end,format!("{:?}",t.class).to_lowercase())}println!();\n }\n}\n'
(fixture_path/'oracle.rs').write_text(oracle_source.replace('ORACLE_SOURCE',json.dumps(str(root/'crates/orbit-pi/src/highlight.rs'))))
subprocess.run([os.environ.get('RUSTC','rustc'),'--edition=2021',str(fixture_path/'oracle.rs'),'-o',str(fixture_path/'oracle')],check=True)

langs=['Rust','Script','Python','Go','C','Java','Ruby','Swift','Json','Yaml','Toml','Shell','Css','Html','Sql','Diff'];tags=['rs','ts','py','go','c','java','rb','swift','json','yaml','toml','sh','css','html','sql','diff']
common=['','let mut count = compute(42) as u32; // note','"escaped \\" quote" tail','/* outer /* inner */ tail(', '/* open', 'close */ let x = 2', '"""open','close""" x()', "'''open", "close''' def answer():", '`open','close` const x = null;', ' @decorator(value) # tail', '#include <a.h>', '  #if FOO', '0xDEAD_u32 0b0101 0o755 0XFF 1_000 2.5e-3f64 4n 1..2 8e+ 3π', 'Self True False None Some Ok Err NaN Infinity NULL nullptr nil', 'Συνάρτηση(世界) Éclair ﬃ x² xⅧ ᢅmark ᶜLower ⒶName ¼', 'a\u2003( b\u0085( c\u001c(', '@name $name :name .class &name #color', 'FOO=1 cargo run --release -p src/main.rs | rg -n "some x" 2>&1 && echo ${HOME} $?', "echo 'literal\\'tail' \"double\\\"tail\" # done", 'cmd >out 2>>log &>file; next ./test ~/file /dev/null -- --flag=value a/b', 'if true; then echo $USER; fi', '${unterminated', '$ $é $a_1 $#@*!$ foo=`command`', '--- a/file', '+++ b/file', '+added', '-removed', '@@ -1 +1 @@', 'diff --git a b', 'index abc..def', ' ordinary', '<!-- open','closed --> <tag key="value">', "'unterminated", '"unterminated', 'select SELECT from FROM null NULL', 'yes no on off ~', 'fn() Function( Type true(', 'x\x00y 42', 'é // 😀', 'a\u0301b Ω𐐀( x𝟎']
for name,tag in zip(langs,tags):
 cases=[]
 block=re.search(r'Lang::'+name+r' => LangSpec \{(.*?)\n        \},',src,re.S).group(1)
 for k in ['keywords','literals','types']:
  match=re.search(k+r': &\[(.*?)\]',block,re.S)
  words=re.findall(r'"([^"\\]*)"',match.group(1)) if match else []
  if words:
   cases += [(k,0,0,' '.join(words)),(k+'_calls',0,0,' '.join(w+'(' for w in words))]
 cases += [('common_'+str(i),0,0,s) for i,s in enumerate(common)]
 for kind,slot in [(1,0)]+[(2,i) for i in range(5)]+[(2,255)]:
  for n,text in enumerate(['', 'closing */ or --> " done\' ` next(', '""" triple\'\'\' rest', 'π😀unterminated']):cases.append((f'carry_{kind}_{slot}_{n}',kind,slot,text))
 p=subprocess.run([str(fixture_path/'oracle')],input=''.join(f'{tag}\t{k}\t{s}\t{text.encode().hex()}\n' for _,k,s,text in cases),text=True,capture_output=True,check=True)
 records=[]
 for case,line in zip(cases,p.stdout.splitlines()):
  ident,kind,slot,text=case;parts=line.split('\t');tokens=[]
  for triple in parts[2:]:
   start,end,cls=triple.split(',');tokens.append({'Start':int(start),'End':int(end),'Class':cls})
  records.append({'Name':ident,'Tag':tag,'Text':text,'Carry':{'Kind':kind,'StringSlot':slot},'Next':{'Kind':int(parts[0]),'StringSlot':int(parts[1])},'Tokens':tokens})
 (root/f'internal/highlight/testdata/{tag}.json').write_text(json.dumps(records,ensure_ascii=False,separators=(',',':'))+'\n')
 print(tag,len(records))
print('source sha256',hashlib.sha256(src.encode()).hexdigest())
