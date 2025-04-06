import os

# 需要排除的目录和文件扩展名
EXCLUDE_DIRS = {'.git', '__pycache__', 'node_modules', 'venv', '.vscode', 'build', 'dist'}
EXCLUDE_EXTENSIONS = {'.md', '.py'}
OUTPUT_FILENAME = 'all_code.md'

# 文件扩展名到 Markdown 语言标识符的映射
# 您可以根据需要扩展这个映射
EXTENSION_MAP = {
    '.py': 'python',
    '.js': 'javascript',
    '.ts': 'typescript',
    '.java': 'java',
    '.c': 'c',
    '.cpp': 'cpp',
    '.cs': 'csharp',
    '.go': 'go',
    '.rb': 'ruby',
    '.php': 'php',
    '.html': 'html',
    '.css': 'css',
    '.scss': 'scss',
    '.sql': 'sql',
    '.sh': 'bash',
    '.yaml': 'yaml',
    '.yml': 'yaml',
    '.json': 'json',
    '.xml': 'xml',
    '.kt': 'kotlin',
    '.swift': 'swift',
    '.rs': 'rust',
    # 添加更多...
}

def get_language_identifier(filename):
    """根据文件扩展名获取 Markdown 语言标识符"""
    _, ext = os.path.splitext(filename)
    return EXTENSION_MAP.get(ext.lower(), '') # 如果找不到映射，则返回空字符串

def combine_code_to_markdown(root_dir, output_file):
    """遍历目录，将代码文件合并到 Markdown 文件中"""
    with open(output_file, 'w', encoding='utf-8') as outfile:
        outfile.write(f"# {os.path.basename(root_dir)} Codebase\n\n") # 添加主标题

        for dirpath, dirnames, filenames in os.walk(root_dir, topdown=True):
            # 修改 dirnames 列表以原地排除目录
            dirnames[:] = [d for d in dirnames if d not in EXCLUDE_DIRS]

            # 过滤掉隐藏目录（以 . 开头）和排除列表中的目录
            rel_dir = os.path.relpath(dirpath, root_dir)
            if rel_dir != '.' and any(part.startswith('.') or part in EXCLUDE_DIRS for part in rel_dir.split(os.sep)):
                continue


            for filename in filenames:
                # 检查文件扩展名是否需要排除
                _, ext = os.path.splitext(filename)
                if ext.lower() in EXCLUDE_EXTENSIONS:
                    continue

                # 检查文件名是否是输出文件本身
                if filename == OUTPUT_FILENAME and dirpath == root_dir:
                    continue

                # 检查隐藏文件（以 . 开头），但保留一些常见的配置文件（可选）
                # if filename.startswith('.'):
                #     continue

                file_path = os.path.join(dirpath, filename)
                relative_path = os.path.relpath(file_path, root_dir)

                try:
                    with open(file_path, 'r', encoding='utf-8', errors='ignore') as infile:
                        content = infile.read()

                    # 写入文件路径作为标题
                    outfile.write(f"## `{relative_path}`\n\n")

                    # 写入代码块
                    lang = get_language_identifier(filename)
                    outfile.write(f"```{lang}\n")
                    outfile.write(content.strip() + "\n") # 去除首尾空白并添加换行符
                    outfile.write("```\n\n")

                except Exception as e:
                    print(f"Error processing file {file_path}: {e}")

if __name__ == "__main__":
    project_root = os.getcwd() # 使用当前工作目录作为项目根目录
    combine_code_to_markdown(project_root, OUTPUT_FILENAME)
    print(f"All code combined into {OUTPUT_FILENAME}")
