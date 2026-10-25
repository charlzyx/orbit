#!/usr/bin/env python3
"""Compile unchanged pinned GitPanel state methods with minimal std-only adapters.

This checks source decisions, not GPUI rendering/dispatch. RUSTC must point at
an already installed toolchain. No downloads, user repositories or gh calls.
"""
import hashlib
import os
import pathlib
import subprocess
import tempfile

ROOT = pathlib.Path(__file__).resolve().parents[3]
SOURCE = ROOT / "crates/orbit-pi/src/git_panel/mod.rs"
EXPECTED = "bbfc71243ee6d2f93eb9cddf3dc8bb0749b677244f478cdce69f2c836bcfef85"
raw = SOURCE.read_bytes()
assert hashlib.sha256(raw).hexdigest() == EXPECTED, "pinned source changed"
source = raw.decode()


def method(signature):
    start = source.index(signature)
    brace = source.index("{", start)
    depth = 1
    end = brace + 1
    while depth:
        depth += (source[end] == "{") - (source[end] == "}")
        end += 1
    return source[start:end]


# The filter/sort is a literal contiguous expression excerpt, not a rewritten
# implementation. The rest of pr_base_popup builds GPUI nodes and is NOT run.
popup = method("fn pr_base_popup(")
start = popup.index("let needle =")
end = popup.index("locals.sort_by_key(|name| name.to_lowercase());")
filter_expr = popup[start:end + len("locals.sort_by_key(|name| name.to_lowercase());")]
adapter = r'''
#![allow(dead_code, unused_variables)]
use std::{path::PathBuf, marker::PhantomData};
mod git {
    #[derive(Clone, Copy, PartialEq)] pub enum RefKind { Branch, Remote, Tag }
    #[derive(Default)] pub struct HistoryFilter;
}
mod gh { #[derive(Default)] pub struct IssueFilter; #[derive(Default)] pub struct PrFilter; }
#[derive(Clone)] struct Ref { name: String, kind: git::RefKind }
struct Context<T>(PhantomData<T>);
impl<T> Context<T> { fn new() -> Self { Self(PhantomData) } }
#[derive(Default)] struct Review;
impl Review { fn set_workspace(&mut self, _:Option<PathBuf>,_: &mut Context<Review>){}
fn set_section_actions(&mut self,_:Option<()>,_: &mut Context<Review>){} }
#[derive(Default)] struct ReviewHandle;
impl ReviewHandle { fn update<F>(&self,_:&mut Context<GitPanel>,f:F)
 where F: FnOnce(&mut Review,&mut Context<Review>) { f(&mut Review,&mut Context::new()) } }
#[derive(Default)] struct Input(String);
impl Input { fn read(&self,_:&mut Context<GitPanel>)->&Self { self } fn text(&self)->&str { &self.0 } }
#[derive(Default)] struct GitPanel {
 workspace:Option<PathBuf>, review:ReviewHandle,
 staged:Vec<()>,unstaged:Vec<()>,history:Vec<()>,graph:Vec<()>,
 remote_web:Option<String>,status:Option<String>,selected_commit:Option<String>,commit_detail:Option<()>,
 history_filter:git::HistoryFilter,file_menu:Option<()>,issues:Vec<()>,issue_detail:Option<()>,
 issue_new_open:bool,issue_generating:bool,issue_filter:gh::IssueFilter,
 pulls:Vec<()>,pr_detail:Option<()>,pr_new_open:bool,pr_generating:bool,pr_filter:gh::PrFilter,
 issue_templates:Vec<()>,pr_templates:Vec<()>,issue_template:Option<usize>,pr_template:Option<usize>,
 template_menu:Option<()>,gh_probed:bool,open:bool,provider:String,model:String,
 refs:Vec<Ref>,branch:Option<String>,branch_filter:Input,pr_new_base:Option<String>,
 issue_new_title:Input,issue_new_body:Input,issue_new_hint:Input,
 pr_new_title:Input,pr_new_body:Input,pr_new_hint:Input,issue_labels:Vec<String>,pr_new_draft:bool,
 refresh_count:usize,
}
impl GitPanel {
 fn refresh_all(&mut self,_:&mut Context<Self>) { self.refresh_count+=1; }
'''
tests = r'''
}
fn entry(name:&str,kind:git::RefKind)->Ref { Ref{name:name.into(),kind} }
#[test] fn default_precedence_and_current_fallback() {
 use git::RefKind::*;
 let mut p=GitPanel::default();
 assert_eq!(p.default_base(),"main");
 p.branch=Some("feature".into()); p.refs=vec![entry("feature",Branch)];
 assert_eq!(p.default_base(),"main");
 p.refs.push(entry("release",Branch)); assert_eq!(p.default_base(),"release");
 p.refs.push(entry("master",Branch)); assert_eq!(p.default_base(),"master");
 p.refs.push(entry("main",Branch)); assert_eq!(p.default_base(),"main");
 p.branch=Some("main".into()); assert_eq!(p.default_base(),"main");
 // Source's preferred-name pass does not check RefKind.
 p.refs=vec![entry("main",Tag),entry("release",Branch)]; assert_eq!(p.default_base(),"main");
 p.refs=vec![entry("remote",Remote),entry("v1",Tag)]; assert_eq!(p.default_base(),"main");
}
#[test] fn literal_local_filter_sort_preserves_source_case_ties() {
 use git::RefKind::*;
 let mut p=GitPanel::default();
 p.refs=vec![entry("zeta",Branch),entry("Beta",Branch),entry("beta",Branch),entry("ALPHA",Branch),entry("Σ",Branch),entry("beta-remote",Remote),entry("beta-tag",Tag)];
 let mut cx=Context::new();
 assert_eq!(p.local_matches(&mut cx),vec!["ALPHA","Beta","beta","zeta","Σ"]);
 p.branch_filter.0="  bEt  ".into(); assert_eq!(p.local_matches(&mut cx),vec!["Beta","beta"]);
 p.branch_filter.0="σ".into(); assert_eq!(p.local_matches(&mut cx),vec!["Σ"]);
 p.branch_filter.0="absent".into(); assert!(p.local_matches(&mut cx).is_empty());
 p.refs=vec![entry("ΟΣ",Branch),entry("İtem",Branch)];
 p.branch_filter.0="ος".into(); assert_eq!(p.local_matches(&mut cx),vec!["ΟΣ"]);
 p.branch_filter.0="i\u{307}".into(); assert_eq!(p.local_matches(&mut cx),vec!["İtem"]);
}
#[test] fn changed_workspace_retains_inputs_and_labels_but_clears_lists_templates() {
 let mut p=GitPanel::default();
 p.workspace=Some("before".into()); p.open=true;
 p.issue_new_title.0="issue title".into();p.issue_new_body.0="issue body".into();p.issue_new_hint.0="issue notes".into();
 p.pr_new_title.0="pull title".into();p.pr_new_body.0="pull body".into();p.pr_new_hint.0="pull notes".into();
 p.issue_labels=vec!["cached-label".into()];p.pr_new_draft=true;p.pr_new_base=Some("release".into());
 p.issues=vec![()];p.pulls=vec![()];p.issue_templates=vec![()];p.pr_templates=vec![()];
 p.issue_new_open=true;p.pr_new_open=true;p.gh_probed=true;
 p.set_context(Some("after".into()),"provider".into(),"model".into(),&mut Context::new());
 assert_eq!(p.issue_new_title.0,"issue title");assert_eq!(p.issue_new_body.0,"issue body");assert_eq!(p.issue_new_hint.0,"issue notes");
 assert_eq!(p.pr_new_title.0,"pull title");assert_eq!(p.pr_new_body.0,"pull body");assert_eq!(p.pr_new_hint.0,"pull notes");
 assert_eq!(p.issue_labels,vec!["cached-label"]);assert!(p.pr_new_draft);assert_eq!(p.pr_new_base.as_deref(),Some("release"));
 assert!(p.issues.is_empty()&&p.pulls.is_empty()&&p.issue_templates.is_empty()&&p.pr_templates.is_empty());
 assert!(!p.issue_new_open&&!p.pr_new_open&&!p.gh_probed);assert_eq!(p.refresh_count,1);
 p.issues.push(());p.issue_new_open=true;
 p.set_context(Some("after".into()),"new-provider".into(),"new-model".into(),&mut Context::new());
 assert_eq!(p.issues.len(),1);assert!(p.issue_new_open);assert_eq!(p.refresh_count,1);assert_eq!(p.provider,"new-provider");
}
'''
program = adapter + method("fn default_base(") + "\n" + method("pub fn set_context(")
program += "\nfn local_matches(&self,cx:&mut Context<Self>)->Vec<String>{\n" + filter_expr + "\nlocals\n}\n" + tests
with tempfile.TemporaryDirectory(prefix="orbit-github-state-oracle-") as directory:
    root = pathlib.Path(directory)
    src = root / "oracle.rs"
    src.write_text(program)
    binary = root / "oracle"
    subprocess.run([os.environ.get("RUSTC", "rustc"), "--edition=2021", "--test", str(src), "-o", str(binary)], check=True)
    print("Pinned GitPanel SHA256:", EXPECTED, flush=True)
    raise SystemExit(subprocess.run([str(binary), "--test-threads=1", "--nocapture"], cwd=root).returncode)
