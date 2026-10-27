# Settings Chinese completion; generated catalogs remain owned by gen_locales.py.
GLOSSARY = {'zh-CN': {'Session retention': '会话保留',
           'Clean up old sessions automatically': '自动清理旧会话',
           'Runs once at launch. Only sessions with no activity in the idle window below are touched.': '启动时执行一次，仅处理超过下方闲置时限未活动的会话。',
           'Archive hides the session but keeps its file; delete removes it from disk.': '归档会隐藏会话但保留文件；删除会从磁盘移除文件。',
           'Idle for': '闲置时长',
           'Measured from the last message in the session, not when it was created.': '从会话的最后一条消息起计算，而非创建时间。',
           'Clean up now': '立即清理',
           'Apply the policy once, right now.': '立即应用一次此策略。',
           'Archived %{count} old sessions': '已归档 %{count} 个旧会话',
           'Deleted %{count} old sessions': '已删除 %{count} 个旧会话',
           '%{count} skipped (in use)': '已跳过 %{count} 个（正在使用）',
           'Temporary files': '临时文件',
           "Orbit's scratch files, kept in its own folder inside the system temp directory.": 'Orbit '
                                                                                              '的临时工作文件，保存在系统临时目录中的专用文件夹内。',
           'Bug-report screenshots, update downloads, and helper scripts are written here. Older `orbit-*` entries are cleaned too.': '问题报告截图、更新下载和辅助脚本写入此处。旧的 '
                                                                                                                                      '`orbit-*` '
                                                                                                                                      '条目也会清理。',
           'Open folder': '打开文件夹',
           'Nothing to clean up.': '没有需要清理的内容。',
           '+%{count} more items': '还有 %{count} 项',
           'Removes these temporary files. Files a live Orbit process still needs are kept.': '移除这些临时文件。正在运行的 '
                                                                                              'Orbit '
                                                                                              '进程仍需使用的文件会保留。',
           'Removed %{count} temporary items (%{size})': '已移除 %{count} 个临时项目（%{size}）',
           '%{count} kept (recent or in use)': '已保留 %{count} 个（最近使用或正在使用）',
           '%{count} could not be removed': '有 %{count} 个无法移除',
           'AI scratch files': 'AI 临时文件',
           "pi's shell output, codemode images, and share/editor working files, written directly into the system temp directory.": 'pi '
                                                                                                                                   '的 '
                                                                                                                                   'shell '
                                                                                                                                   '输出、codemode '
                                                                                                                                   '图片以及共享/编辑器工作文件，直接写入系统临时目录。',
           "pi writes these files directly into the temp root, outside Orbit's folder. Files used in the last hour are kept.": 'pi '
                                                                                                                               '将这些文件直接写入临时根目录，而非 '
                                                                                                                               'Orbit '
                                                                                                                               '文件夹。最近一小时使用过的文件会保留。',
           'No AI scratch files.': '没有 AI 临时文件。',
           "Removes pi's scratch files last used more than an hour ago. Recent files are kept.": '移除 pi '
                                                                                                 '超过一小时未使用的临时文件。最近使用的文件会保留。',
           'Show the sessions sidebar. Toggle it from the top bar or with %{shortcut}.': '显示会话侧边栏。可从顶部栏或使用 '
                                                                                         '%{shortcut} 切换。',
           'Keyboard chords for the workbench, the composer, and the transcript.': '工作台、输入框和对话记录的键盘快捷键。',
           'All providers': '所有提供商',
           'Active model': '当前模型',
           'The model the current session uses.': '当前会话使用的模型。',
           'Update model list': '更新模型列表',
           'Model list updated': '模型列表已更新',
           'Could not update the model list: %{detail}': '无法更新模型列表：%{detail}',
           'Base URL': '基础 URL',
           'Add to favorites': '添加到收藏',
           'Remove from favorites': '从收藏中移除',
           'Model Context Protocol servers Orbit manages for the agent.': 'Orbit 为智能体管理的模型上下文协议服务器。',
           'Search MCP servers…': '搜索 MCP 服务器…',
           'Add MCP Server': '添加 MCP 服务器',
           'Refresh servers': '刷新服务器',
           'Checking connections…': '正在检查连接…',
           'Checked %{seconds}s ago': '%{seconds} 秒前检查',
           'Not checked yet': '尚未检查',
           'No MCP servers configured.': '尚未配置 MCP 服务器。',
           '↑↓ Navigate · ⏎ Edit · Space Enable/Disable': '↑↓ 导航 · ⏎ 编辑 · 空格 启用/禁用',
           'Connect an MCP server to give your agent additional tools.': '连接 MCP 服务器，为智能体提供更多工具。',
           'Try a server name, a command, or a URL.': '试试服务器名称、命令或 URL。',
           'Sign-in required': '需要登录',
           'Connection failed': '连接失败',
           'This Project': '当前项目',
           'Available across all Orbit projects.': '可用于所有 Orbit 项目。',
           'Available only in the current project.': '仅可用于当前项目。',
           'Streamable HTTP': '流式 HTTP',
           'This MCP server runs a local process on this computer.': '此 MCP 服务器在本机运行一个本地进程。',
           'This MCP server communicates with an external service.': '此 MCP 服务器与外部服务通信。',
           'Callable from codemode scripts, not declared to the model.': '可从 codemode 脚本调用，不向模型声明。',
           'Like codemode, but not listed in the codemode description.': '与 codemode 类似，但不列入 codemode 描述。',
           'The model loads the tools on demand with tool_search.': '模型通过 tool_search 按需加载工具。',
           'Declared to the model like built-in tools.': '像内置工具一样向模型声明。',
           'Registered but not callable.': '已注册，但不可调用。',
           'Letters, digits, underscores, and dashes only.': '仅允许字母、数字、下划线和连字符。',
           'What this server is for': '此服务器的用途',
           "Shown only in Orbit; never written to the agent's configuration.": '仅在 Orbit 中显示，不会写入智能体配置。',
           'A single executable, not a shell string.': '单个可执行文件，而非 shell 命令字符串。',
           'One argument per line.': '每行一个参数。',
           'Environment variables': '环境变量',
           "One KEY=value per line. Literal values are stored in Orbit's secret store and written as ${NAME} references.": '每行一个 '
                                                                                                                           'KEY=value。字面值保存在 '
                                                                                                                           'Orbit '
                                                                                                                           '的私密存储中，并写为 '
                                                                                                                           '${NAME} '
                                                                                                                           '引用。',
           'A streamable HTTP endpoint (https://…/mcp). The legacy SSE transport is not supported.': '流式 '
                                                                                                     'HTTP '
                                                                                                     '端点（https://…/mcp）。不支持旧版 '
                                                                                                     'SSE '
                                                                                                     '传输。',
           "One Name: value per line. Literal values are stored in Orbit's secret store and written as ${NAME} references.": '每行一个 '
                                                                                                                             'Name: '
                                                                                                                             'value。字面值保存在 '
                                                                                                                             'Orbit '
                                                                                                                             '的私密存储中，并写为 '
                                                                                                                             '${NAME} '
                                                                                                                             '引用。',
           'The agent connects and calls the tools; Orbit manages the configuration.': '智能体负责连接和调用工具；Orbit '
                                                                                       '负责管理配置。',
           'Test connection': '测试连接',
           'Edit MCP Server': '编辑 MCP 服务器',
           'Configuration problem': '配置问题',
           "The project's MCP configuration is ignored by the standalone check until the project is trusted. Orbit sessions run with project trust, so these servers still load here.": '项目获得信任前，独立检查会忽略项目的 '
                                                                                                                                                                                        'MCP '
                                                                                                                                                                                        '配置。Orbit '
                                                                                                                                                                                        '会话以信任项目的方式运行，因此仍会在此加载这些服务器。',
           'MCP configuration changed on disk': '磁盘上的 MCP 配置已更改',
           'The files were edited outside Orbit. The list below shows the current content; apply the changes to the running session when ready.': '文件已在 '
                                                                                                                                                  'Orbit '
                                                                                                                                                  '外部编辑。下方列表显示当前内容；准备好后可将更改应用到运行中的会话。',
           'Configuration changed': '配置已更改',
           'The agent needs to restart to apply the changes. Your current session will be restored.': '智能体需要重启才能应用更改。当前会话将恢复。',
           'Apply now': '立即应用',
           'Applying MCP configuration…': '正在应用 MCP 配置…',
           'MCP configuration applied — restarting the agent and restoring the session.': 'MCP 配置已应用 — '
                                                                                          '正在重启智能体并恢复会话。',
           'MCP changes will apply when the current run finishes.': '当前运行结束后将应用 MCP 更改。',
           'Reconnecting MCP servers…': '正在重新连接 MCP 服务器…',
           'Connected to %{name} — %{count} tool(s) available.': '已连接到 %{name} — 有 %{count} 个可用工具。',
           'Could not connect to %{name}. %{detail}': '无法连接到 %{name}。%{detail}',
           'No details were reported. Check the command or URL and try again.': '未返回详细信息。请检查命令或 URL 后重试。',
           'Allow MCP · %{server} · %{tool}?': '允许 MCP · %{server} · %{tool}？',
           'MCP: Open Settings': 'MCP：打开设置',
           'MCP: Refresh Servers': 'MCP：刷新服务器',
           'MCP: Reconnect All': 'MCP：全部重新连接',
           'Invalid MCP configuration: %{detail}': 'MCP 配置无效：%{detail}',
           'MCP server “%{name}” was not found.': '未找到 MCP 服务器“%{name}”。',
           "Unable to start the MCP server's process: %{detail}": '无法启动 MCP 服务器进程：%{detail}',
           'Unable to connect to the MCP server: %{detail}': '无法连接到 MCP 服务器：%{detail}',
           'Authentication failed: %{detail}': '身份验证失败：%{detail}',
           'The MCP operation timed out: %{detail}': 'MCP 操作超时：%{detail}',
           "Could not update the agent's MCP configuration: %{detail}": '无法更新智能体的 MCP 配置：%{detail}',
           'The agent is not available, so MCP servers cannot be checked.': '智能体不可用，无法检查 MCP 服务器。',
           'Unsupported MCP transport: %{detail}': '不支持的 MCP 传输：%{detail}',
           'MCP configuration I/O failed: %{detail}': 'MCP 配置读写失败：%{detail}',
           'A server name is required.': '请填写服务器名称。',
           'Server names may only contain letters, digits, “_”, and “-”.': '服务器名称只能包含字母、数字、“_”和“-”。',
           'A stdio server needs a command.': 'stdio 服务器需要指定命令。',
           'The command contains an invalid character.': '命令包含无效字符。',
           'The working directory contains an invalid character.': '工作目录包含无效字符。',
           'Timeout must be a positive number of seconds.': '超时必须为正数，单位为秒。',
           'Enter a valid URL.': '请输入有效的 URL。',
           'The URL must start with http:// or https://.': 'URL 必须以 http:// 或 https:// 开头。',
           'The OAuth callback URL is not valid.': 'OAuth 回调 URL 无效。',
           'The OAuth callback URL must be an http loopback address (localhost, 127.0.0.1, or [::1]).': 'OAuth '
                                                                                                        '回调 '
                                                                                                        'URL '
                                                                                                        '必须是 '
                                                                                                        'http '
                                                                                                        '回环地址（localhost、127.0.0.1 '
                                                                                                        '或 '
                                                                                                        '[::1]）。',
           'Could not read %{path}: %{detail}': '无法读取 %{path}：%{detail}',
           '%{path} is not valid JSON: %{detail}': '%{path} 不是有效的 JSON：%{detail}',
           '%{path} must contain a JSON object.': '%{path} 必须包含 JSON 对象。',
           '%{path}: “mcpServers” must be an object.': '%{path}：“mcpServers”必须是对象。',
           '%{path}: “autoEnableCodemode” must be true or false.': '%{path}：“autoEnableCodemode”必须为 true 或 '
                                                                   'false。',
           '%{path}: server “%{name}” was skipped — %{detail}': '%{path}：已跳过服务器“%{name}” — %{detail}',
           'the server entry must be an object': '服务器条目必须是对象',
           '“type” must be a string': '“type”必须是字符串',
           'the legacy SSE transport is not supported; use the streamable HTTP URL': '不支持旧版 SSE 传输；请使用流式 '
                                                                                     'HTTP URL',
           '“enabled” must be true or false': '“enabled”必须为 true 或 false',
           '“exposure” must be a string': '“exposure”必须是字符串',
           '“%{mode}” is not a valid exposure mode': '“%{mode}”不是有效的暴露模式',
           'toolExposure.%{tool} must be a string': '应将 toolExposure.%{tool} 设为字符串',
           'toolExposure.%{tool} is not a valid exposure mode (“%{mode}”)': 'toolExposure.%{tool} '
                                                                            '的暴露模式无效（“%{mode}”）',
           '“toolExposure” must map tool names to exposure modes': '“toolExposure”必须将工具名称映射到暴露模式',
           '“timeout” must be a number': '“timeout”必须是数字',
           '“type” does not match the HTTP server': '“type”与 HTTP 服务器不匹配',
           '“type” does not match the stdio server': '“type”与 stdio 服务器不匹配',
           '“url” must be a string': '“url”必须是字符串',
           '“command” must be a string': '“command”必须是字符串',
           '“args” must be an array': '“args”必须是数组',
           'every argument must be a string': '每个参数必须是字符串',
           '“cwd” must be a string': '“cwd”必须是字符串',
           'either “command” (stdio) or “url” (streamable HTTP) is required': '必须指定“command”（stdio）或“url”（流式 '
                                                                              'HTTP）',
           '“oauth” must be an object': '“oauth”必须是对象',
           'oauth.%{field} must be a string': 'oauth.%{field} 必须是字符串',
           'oauth.callbackPort must be a valid port number': 'oauth.callbackPort 必须是有效的端口号',
           '“%{field}” must map names to strings': '“%{field}”必须将名称映射到字符串',
           '%{field}.%{key} must be a string': '%{field}.%{key} 必须是字符串',
           'Could not update %{path}: %{detail}': '无法更新 %{path}：%{detail}',
           'line %{line} must be KEY=value': '第 %{line} 行必须为 KEY=value',
           'line %{line} must be Name: value': '第 %{line} 行必须为 Name: value',
           'A project-scoped MCP server needs an open project.': '项目范围的 MCP 服务器需要已打开的项目。',
           'Could not run the agent to check MCP servers: %{detail}': '无法运行智能体来检查 MCP 服务器：%{detail}',
           'Checking MCP servers timed out.': '检查 MCP 服务器超时。',
           'The agent did not report any MCP status.': '智能体没有返回 MCP 状态。',
           'Could not check MCP servers: %{detail}': '无法检查 MCP 服务器：%{detail}',
           "Could not read the agent's MCP status: %{detail}": '无法读取智能体的 MCP 状态：%{detail}',
           'Could not restart the agent to apply MCP changes: %{error}': '无法重启智能体以应用 MCP 更改：%{error}',
           'Codemode is disabled globally': '已全局禁用 codemode',
           'Some servers below expose tools through codemode. With “autoEnableCodemode” set to false in mcp.json, those tools cannot be called until the setting is re-enabled or the servers use a different exposure.': '下方部分服务器通过 '
                                                                                                                                                                                                                          'codemode '
                                                                                                                                                                                                                          '暴露工具。当 '
                                                                                                                                                                                                                          'mcp.json '
                                                                                                                                                                                                                          '中的“autoEnableCodemode”为 '
                                                                                                                                                                                                                          'false '
                                                                                                                                                                                                                          '时，这些工具无法调用，直到重新启用该设置或改用其他暴露方式。',
           'Check that the command is installed and available to the agent.': '请检查该命令是否已安装且智能体可以访问。',
           'Check the configured authentication credentials.': '请检查配置的身份验证凭据。',
           'The server may be slow to start; raise its timeout in mcp.json if needed.': '服务器启动可能较慢；如有需要，可提高 '
                                                                                        'mcp.json 中的超时设置。',
           'Check the URL and the network connection.': '请检查 URL 和网络连接。',
           'MCP %{connected}/%{total} · sign-in needed': 'MCP %{connected}/%{total} · 需要登录',
           'Sign out': '退出登录',
           'Waiting for browser…': '正在等待浏览器…',
           'Signing in to %{name} — finish in your browser.': '正在登录 %{name} — 请在浏览器中完成。',
           'Cancelling sign-in to %{name}…': '正在取消登录 %{name}…',
           'Sign-in to %{name} cancelled.': '已取消登录 %{name}。',
           'Signed in to %{name}.': '已登录 %{name}。',
           'Signed out of %{name}.': '已退出 %{name}。',
           'OAuth client settings (optional)': 'OAuth 客户端设置（可选）',
           'Only for pre-registered clients. Most servers need nothing here — save the URL, then use Sign in.': '仅用于预先注册的客户端。多数服务器无需填写 '
                                                                                                                '— '
                                                                                                                '保存 '
                                                                                                                'URL '
                                                                                                                '后使用“登录”即可。',
           'Client ID': '客户端 ID',
           'Client secret': '客户端密钥',
           "Stored in Orbit's secret store; written as a ${NAME} reference.": '保存在 Orbit 的私密存储中，以 ${NAME} '
                                                                              '引用写入。',
           'client secret': '客户端密钥',
           'Callback port': '回调端口',
           'Fixed loopback port for clients registered with a redirect URI.': '为使用重定向 URI 注册的客户端指定固定回环端口。',
           'Callback URL': '回调 URL',
           'Must be an http loopback URI (localhost, 127.0.0.1, or [::1]).': '必须是 http 回环 '
                                                                             'URI（localhost、127.0.0.1 或 '
                                                                             '[::1]）。',
           'Sign-in to %{name} timed out waiting for the browser.': '等待浏览器完成 %{name} 登录时超时。',
           'Sign-out for %{name} timed out.': '退出 %{name} 登录超时。',
           'Could not sign in to %{name}. %{detail}': '无法登录 %{name}。%{detail}',
           'Could not sign out of %{name}. %{detail}': '无法退出 %{name} 登录。%{detail}',
           'Open MCP settings': '打开 MCP 设置',
           'Action': '操作',
           'Archive': '归档',
           'Delete': '删除',
           '%{days} days': '%{days} 天',
           '%{count} failed': '%{count} 项失败',
           'Location': '位置',
           'File': '文件',
           'Folder': '文件夹',
           'kept': '已保留',
           '%{count} items · %{size}': '%{count} 项 · %{size}',
           'Shortcuts': '快捷键',
           'Privacy': '隐私',
           'Clear': '清除',
           '%{count} recent': '最近使用 %{count} 项',
           'Recents': '最近使用',
           'Reload': '重新加载',
           '%{count} server(s)': '%{count} 个服务器',
           '· disabled': '· 已禁用',
           'Disabled': '已禁用',
           '%{count} tool(s)': '%{count} 个工具',
           '%{count} resource(s)': '%{count} 个资源',
           'Connecting…': '正在连接…',
           'Disconnected': '已断开连接',
           'Scope': '作用范围',
           'Global': '全局',
           'Codemode (deferred)': 'Codemode（延迟加载）',
           'Deferred': '延迟加载',
           'Direct': '直接暴露',
           'Hidden': '隐藏',
           'Description': '描述',
           'Headers': '请求头',
           'Save': '保存',
           'Saved %{name}.': '已保存 %{name}。',
           'Removed %{name}.': '已移除 %{name}。',
           '%{name} enabled.': '已启用 %{name}。',
           '%{name} disabled.': '已禁用 %{name}。',
           'MCP: Enable %{name}': 'MCP：启用 %{name}',
           'MCP: Disable %{name}': 'MCP：禁用 %{name}',
           'MCP: Reconnect %{name}': 'MCP：重新连接 %{name}',
           'MCP: Test %{name}': 'MCP：测试 %{name}',
           ' · project: %{path}': ' · 项目：%{path}',
           'MCP %{connected}/%{total} · %{failed} failed': 'MCP %{connected}/%{total} · %{failed} 个失败',
           '%{count} tools': '%{count} 个工具',
           '%{connected}/%{total} connected': '已连接 %{connected}/%{total}',
           'How Orbit stores and cleans up local data.': 'Orbit 如何存储和清理本地数据。',
           'Workbench': '工作台',
           'Composer': '输入框',
           'Transcript': '对话记录',
           'Application': '应用程序',
           'Send Message': '发送消息',
           'Steer Running Turn': '引导当前运行轮次',
           'Insert New Line': '插入换行',
           'Accept Suggestion': '接受建议',
           'Stop Run': '停止运行',
           'Previous User Turn': '上一用户轮次',
           'Next User Turn': '下一用户轮次',
           'Copy Last Response': '复制上一条回复',
           'Exposure': '暴露方式'}}
