#![allow(dead_code)]
#[derive(Clone)]
enum Block {
    Paragraph(Vec<String>),
    Heading(u8, String),
    Code(Option<String>, Vec<String>),
    List(Vec<ListItem>),
    Quote(Vec<String>),
    /// GitHub-style alert (`> [!NOTE]` …): a labeled, tinted callout.
    Alert(AlertKind, Vec<String>),
    Rule,
    Table {
        header: Vec<String>,
        rows: Vec<Vec<String>>,
        aligns: Vec<TableAlign>,
    },
    /// A standalone image: a Markdown `![alt](url)` line or an HTML
    /// `<img src="…">` tag. GitHub issue screenshots usually land on their
    /// own line, so the block model is enough to render them inline.
    Image {
        alt: String,
        url: String,
    },
}

/// GFM column alignment parsed from a table's delimiter row.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
enum TableAlign {
    #[default]
    Left,
    Center,
    Right,
}


#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum AlertKind {
    Note,
    Tip,
    Important,
    Warning,
    Caution,
}


#[derive(Clone)]
struct ListItem {
    depth: usize,
    ordered: bool,
    number: u64,
    /// GFM task-list state: `Some(true)` checked, `Some(false)` open, `None`
    /// a plain item. Task items drop the bullet/number for a checkbox.
    checked: Option<bool>,
    text: String,
}

fn flush_paragraph(paragraph: &mut Vec<String>, blocks: &mut Vec<Block>) {
    if !paragraph.is_empty() {
        blocks.push(Block::Paragraph(std::mem::take(paragraph)));
    }
}

fn heading_line(trimmed: &str) -> Option<(u8, String)> {
    let level = trimmed.chars().take_while(|&c| c == '#').count();
    if (1..=6).contains(&level) {
        if let Some(body) = trimmed[level..].strip_prefix(' ') {
            let body = body.trim();
            if !body.is_empty() {
                // Keep the real level; the renderer gives h4+ a smaller step
                // rather than flattening everything past h3 to one size.
                return Some((level as u8, body.to_string()));
            }
        }
    }
    None
}

fn is_rule(trimmed: &str) -> bool {
    let compact: String = trimmed.chars().filter(|c| !c.is_whitespace()).collect();
    match compact.chars().next() {
        Some(first @ ('-' | '*' | '_')) => {
            compact.len() >= 3 && compact.chars().all(|c| c == first)
        }
        _ => false,
    }
}

/// Parse a line that is solely an image into `(alt, url)`.
///
/// Handles Markdown `![alt](url "title")` (with optional `<…>` wrapping) and
/// the HTML `<img src="…">` form GitHub sometimes stores. Returns `None` for
/// a line that merely contains an image among other text, so prose is never
/// split mid-sentence.
fn image_line(trimmed: &str) -> Option<(String, String)> {
    let t = trimmed.trim();
    if let Some(rest) = t.strip_prefix("![") {
        let close = rest.find("](")?;
        let alt = rest[..close].to_string();
        let after = &rest[close + 2..];
        let end = after.find(')')?;
        // A title follows the URL after whitespace; the URL itself may be
        // wrapped in angle brackets (Markdown's escape for spaces).
        let raw = after[..end].trim();
        let url = raw
            .strip_prefix('<')
            .and_then(|rest| rest.split('>').next())
            .unwrap_or_else(|| raw.split_whitespace().next().unwrap_or(""));
        if !url.is_empty() {
            return Some((alt, normalize_image_url(url)));
        }
        return None;
    }
    if t.starts_with("<img") {
        let src = html_attr(t, "src")?;
        if !src.is_empty() {
            return Some((
                html_attr(t, "alt").unwrap_or_default(),
                normalize_image_url(&src),
            ));
        }
    }
    None
}

/// Rewrite GitHub's signed `private-user-images` CDN URL to the durable
/// `user-attachments` URL for the same asset.
///
/// A screenshot pasted into a comment is stored as
/// `private-user-images.githubusercontent.com/<user>/<asset>-<uuid>.<ext>?jwt=…`.
/// That JWT expires within minutes, and GitHub re-signs the URL only when it
/// renders the page itself — so the stored URL 404s for Orbit and the reply
/// shows the alt-text fallback. The durable URL is the same asset and never
/// expires: it is the form the issue/PR body itself uses.
fn normalize_image_url(url: &str) -> String {
    const SIGNED_HOST: &str = "https://private-user-images.githubusercontent.com/";
    let Some(rest) = url.strip_prefix(SIGNED_HOST) else {
        return url.to_string();
    };
    let path = rest.split(['?', '#']).next().unwrap_or(rest);
    let file = path.rsplit('/').next().unwrap_or(path);
    // `<asset id>-<uuid>.<ext>`: the UUID half names the durable asset.
    let Some((_, stem)) = file.split_once('-') else {
        return url.to_string();
    };
    let uuid = stem.rsplit_once('.').map_or(stem, |(stem, _)| stem);
    if !is_uuid(uuid) {
        return url.to_string();
    }
    format!("https://github.com/user-attachments/assets/{uuid}")
}

/// `8-4-4-4-12` hex, the shape of GitHub's asset IDs.
fn is_uuid(value: &str) -> bool {
    let bytes = value.as_bytes();
    bytes.len() == 36
        && bytes.iter().enumerate().all(|(ix, byte)| match ix {
            8 | 13 | 18 | 23 => *byte == b'-',
            _ => byte.is_ascii_hexdigit(),
        })
}

/// Read `name="value"` (single or double quoted) from a tag's source text.
fn html_attr(tag: &str, name: &str) -> Option<String> {
    let lower = tag.to_ascii_lowercase();
    let key = format!("{name}=");
    let start = lower.find(&key)? + key.len();
    let rest = &tag[start..];
    let quote = rest.chars().next()?;
    if quote != '"' && quote != '\'' {
        return None;
    }
    let rest = &rest[1..];
    let end = rest.find(quote)?;
    Some(rest[..end].to_string())
}

fn list_item_line(line: &str) -> Option<(usize, bool, u64, Option<bool>, String)> {
    let indent = line.len() - line.trim_start().len();
    let t = line.trim_start();
    let (ordered, number, body) = if let Some(rest) = t
        .strip_prefix("- ")
        .or_else(|| t.strip_prefix("* "))
        .or_else(|| t.strip_prefix("+ "))
    {
        (false, 0u64, rest)
    } else {
        let digits = t.chars().take_while(|c| c.is_ascii_digit()).count();
        if digits == 0 {
            return None;
        }
        let after = t[digits..]
            .strip_prefix(". ")
            .or_else(|| t[digits..].strip_prefix(") "))?;
        (true, t[..digits].parse().ok()?, after)
    };
    let (checked, body) = match task_status(body) {
        Some((checked, rest)) => (Some(checked), rest),
        None => (None, body),
    };
    Some((
        indent.div_ceil(2).min(2),
        ordered,
        number,
        checked,
        body.to_string(),
    ))
}

/// Strip a leading GFM task marker (`[ ] `, `[x] `, `[X] `) from an item's
/// body, returning its state and the remaining text.
fn task_status(body: &str) -> Option<(bool, &str)> {
    let rest = body.strip_prefix('[')?;
    let mut chars = rest.chars();
    let checked = match chars.next()? {
        ' ' => false,
        'x' | 'X' => true,
        _ => return None,
    };
    let after = &rest[1..];
    let after = after.strip_prefix(']')?;
    Some((checked, after.strip_prefix(' ').unwrap_or(after)))
}

fn split_table_row(line: &str) -> Vec<String> {
    let s = line.trim();
    let s = s.strip_prefix('|').unwrap_or(s);
    let s = s.strip_suffix('|').unwrap_or(s);
    s.split('|').map(|cell| cell.trim().to_string()).collect()
}

/// Parse a GFM delimiter row (`| --- | :--: | ---: |`) into per-column
/// alignment; `None` when the line is not a delimiter row.
fn table_aligns(trimmed: &str) -> Option<Vec<TableAlign>> {
    if !trimmed.contains('-') || !trimmed.contains('|') {
        return None;
    }
    if !trimmed.chars().all(|c| matches!(c, '|' | '-' | ':' | ' ')) {
        return None;
    }
    let cells = split_table_row(trimmed);
    if cells.is_empty() {
        return None;
    }
    Some(
        cells
            .iter()
            .map(|cell| {
                let left = cell.starts_with(':');
                let right = cell.ends_with(':');
                match (left, right) {
                    (true, true) => TableAlign::Center,
                    (false, true) => TableAlign::Right,
                    _ => TableAlign::Left,
                }
            })
            .collect(),
    )
}

/// GitHub alert marker on the first quoted line, e.g. `[!WARNING]`.
fn alert_marker(body: &[String]) -> Option<AlertKind> {
    let first = body.first()?.trim();
    let inner = first.strip_prefix("[!")?.strip_suffix(']')?;
    Some(match inner.trim().to_ascii_uppercase().as_str() {
        "NOTE" => AlertKind::Note,
        "TIP" => AlertKind::Tip,
        "IMPORTANT" => AlertKind::Important,
        "WARNING" => AlertKind::Warning,
        "CAUTION" => AlertKind::Caution,
        _ => return None,
    })
}

fn parse_blocks(text: &str) -> Vec<Block> {
    let mut blocks: Vec<Block> = Vec::new();
    let mut paragraph: Vec<String> = Vec::new();
    let lines: Vec<&str> = text.lines().collect();
    let mut i = 0;

    while i < lines.len() {
        let line = lines[i];
        let trimmed = line.trim_start();

        if trimmed.starts_with("```") || trimmed.starts_with("~~~") {
            flush_paragraph(&mut paragraph, &mut blocks);
            let marker = &trimmed[..trimmed.len().min(3)];
            // The fence info string's first token is the language hint
            // (```rust, ```text, …) — it becomes the code block's chip.
            let language = trimmed[3..]
                .split_whitespace()
                .next()
                .filter(|lang| !lang.is_empty())
                .map(str::to_string);
            let mut body = Vec::new();
            i += 1;
            while i < lines.len() && !lines[i].trim_start().starts_with(marker) {
                body.push(lines[i].to_string());
                i += 1;
            }
            i += 1; // closing fence (or EOF)
            blocks.push(Block::Code(language, body));
            continue;
        }

        if trimmed.is_empty() {
            flush_paragraph(&mut paragraph, &mut blocks);
            i += 1;
            continue;
        }

        if let Some((level, body)) = heading_line(trimmed) {
            flush_paragraph(&mut paragraph, &mut blocks);
            blocks.push(Block::Heading(level, body));
            i += 1;
            continue;
        }

        if is_rule(trimmed) {
            flush_paragraph(&mut paragraph, &mut blocks);
            blocks.push(Block::Rule);
            i += 1;
            continue;
        }

        // A line that is only an image becomes its own block, so a screenshot
        // pasted under a heading actually paints instead of rendering as a
        // stray `!` and a link.
        if let Some((alt, url)) = image_line(trimmed) {
            flush_paragraph(&mut paragraph, &mut blocks);
            blocks.push(Block::Image { alt, url });
            i += 1;
            continue;
        }

        if trimmed.contains('|') {
            if let Some(aligns) = lines
                .get(i + 1)
                .and_then(|separator| table_aligns(separator.trim()))
            {
                flush_paragraph(&mut paragraph, &mut blocks);
                let header = split_table_row(trimmed);
                i += 2;
                let mut rows = Vec::new();
                while i < lines.len() {
                    let row_line = lines[i].trim();
                    if row_line.is_empty() || !row_line.contains('|') {
                        break;
                    }
                    rows.push(split_table_row(row_line));
                    i += 1;
                }
                blocks.push(Block::Table {
                    header,
                    rows,
                    aligns,
                });
                continue;
            }
        }

        if trimmed.starts_with('>') {
            flush_paragraph(&mut paragraph, &mut blocks);
            let mut body = Vec::new();
            while i < lines.len() {
                match lines[i].trim_start().strip_prefix('>') {
                    Some(rest) => {
                        body.push(rest.strip_prefix(' ').unwrap_or(rest).to_string());
                        i += 1;
                    }
                    None => break,
                }
            }
            match alert_marker(&body) {
                Some(kind) => {
                    body.remove(0);
                    // Drop the blank separator line GitHub allows after the
                    // marker so the callout body starts on real content.
                    while body.first().is_some_and(|line| line.trim().is_empty()) {
                        body.remove(0);
                    }
                    blocks.push(Block::Alert(kind, body));
                }
                None => blocks.push(Block::Quote(body)),
            }
            continue;
        }

        if let Some((depth, ordered, number, checked, body)) = list_item_line(line) {
            flush_paragraph(&mut paragraph, &mut blocks);
            let mut items = Vec::new();
            let mut pending = ListItem {
                depth,
                ordered,
                number,
                checked,
                text: body,
            };
            i += 1;
            while i < lines.len() {
                let l = lines[i];
                if l.trim().is_empty() {
                    // A blank line keeps the list open only when another
                    // item follows (loose list).
                    let mut look = i + 1;
                    while look < lines.len() && lines[look].trim().is_empty() {
                        look += 1;
                    }
                    if look < lines.len() && list_item_line(lines[look]).is_some() {
                        i = look;
                        continue;
                    }
                    break;
                }
                if let Some((depth, ordered, number, checked, body)) = list_item_line(l) {
                    // A marker-type change starts a new list.
                    if ordered != pending.ordered {
                        break;
                    }
                    items.push(std::mem::replace(
                        &mut pending,
                        ListItem {
                            depth,
                            ordered,
                            number,
                            checked,
                            text: body,
                        },
                    ));
                    i += 1;
                } else if l.starts_with("  ") || l.starts_with('\t') {
                    // Indented continuation of the pending item.
                    pending.text.push(' ');
                    pending.text.push_str(l.trim());
                    i += 1;
                } else {
                    break;
                }
            }
            items.push(pending);
            blocks.push(Block::List(items));
            continue;
        }

        paragraph.push(line.trim().to_string());
        i += 1;
    }
    flush_paragraph(&mut paragraph, &mut blocks);
    blocks
}


fn main() {
 let input = "| A | B |\n| --- | --- |\n| one | two | [lost](https://example.invalid/extra) |\n";
 let blocks = parse_blocks(input);
 match &blocks[0] {
  Block::Table { header, rows, .. } => {
   assert_eq!(header.len(),2); assert_eq!(rows.len(),1); assert_eq!(rows[0].len(),3);
   assert_eq!(rows[0][2], "[lost](https://example.invalid/extra)");
   println!("header={:?}; rows={:?}; retained cells={}",header,rows,header.len()+rows[0].len());
  }, _ => panic!("not a table")
 }
}
