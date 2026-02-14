#!/usr/bin/env python3
"""
构建脚本，用于将前端静态资源打包到Go可执行文件中

功能：
1. 检查前端构建文件是否存在
2. 创建static目录结构
3. 复制前端构建文件到正确的位置
4. 多平台交叉编译Go应用
5. 清理临时文件
"""

import argparse
import os
import platform
import shutil
import subprocess
import sys

# 定义路径
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
GOEND_DIR = os.path.dirname(SCRIPT_DIR)
PROJECT_ROOT = os.path.dirname(GOEND_DIR)
API_DIR = os.path.join(GOEND_DIR, "cmd", "api")
STATIC_DIR = os.path.join(API_DIR, "static")
FRONTEND_DIST = os.path.join(PROJECT_ROOT, "frontend", "dist")

# Release目录配置
RELEASE_DIR = os.path.join(PROJECT_ROOT, "release")
CONFIG_FILE = os.path.join(PROJECT_ROOT, "config.yml")

# MSYS2 CGO工具链配置
MSYS2_UCRT64_PATH = r"D:\CodeHub\msys64\ucrt64"

# CGO工具链配置 - Windows本地编译使用MSYS2 UCRT64
CGO_TOOLCHAINS = {
    "windows-amd64": {
        "cc": "gcc",
        "cxx": "g++",
        "msys2_path": MSYS2_UCRT64_PATH,
        "description": "Windows x64 (使用MSYS2 UCRT64 GCC)"
    },
    "windows-arm64": {
        "cc": "gcc",
        "cxx": "g++",
        "msys2_path": MSYS2_UCRT64_PATH,
        "description": "Windows ARM64 (使用MSYS2 UCRT64 GCC)"
    },
    "linux-amd64": {
        "cc": "x86_64-linux-gnu-gcc",
        "cxx": "x86_64-linux-gnu-g++",
        "msys2_path": None,
        "description": "Linux x64 (需要交叉编译工具链)"
    },
    "linux-arm64": {
        "cc": "aarch64-linux-gnu-gcc",
        "cxx": "aarch64-linux-gnu-g++",
        "msys2_path": None,
        "description": "Linux ARM64 (需要交叉编译工具链)"
    },
    "darwin-amd64": {
        "cc": "o64-clang",
        "cxx": "o64-clang++",
        "msys2_path": None,
        "description": "macOS Intel (需要osxcross)"
    },
    "darwin-arm64": {
        "cc": "oa64-clang",
        "cxx": "oa64-clang++",
        "msys2_path": None,
        "description": "macOS Apple Silicon (需要osxcross)"
    },
}

# 输出颜色
GREEN = "\033[92m"
YELLOW = "\033[93m"
RED = "\033[91m"
RESET = "\033[0m"

# 支持的平台配置
PLATFORMS = {
    "windows-amd64": {"goos": "windows", "goarch": "amd64", "ext": ".exe"},
    "windows-arm64": {"goos": "windows", "goarch": "arm64", "ext": ".exe"},
    "linux-amd64": {"goos": "linux", "goarch": "amd64", "ext": ""},
    "linux-arm64": {"goos": "linux", "goarch": "arm64", "ext": ""},
    "darwin-amd64": {"goos": "darwin", "goarch": "amd64", "ext": ""},
    "darwin-arm64": {"goos": "darwin", "goarch": "arm64", "ext": ""},
}

def print_info(message):
    """打印信息消息"""
    print(f"{GREEN}[INFO]{RESET} {message}")

def print_warning(message):
    """打印警告消息"""
    print(f"{YELLOW}[WARNING]{RESET} {message}")

def print_error(message):
    """打印错误消息"""
    print(f"{RED}[ERROR]{RESET} {message}")

def build_frontend():
    """构建前端应用"""
    print_info("开始构建前端应用...")

    frontend_dir = os.path.join(PROJECT_ROOT, "frontend")

    try:
        # 切换到前端目录并执行yarn build
        # Windows下yarn可能是.ps1脚本，需要使用shell=True才能正确执行
        result = subprocess.run(
            "yarn build",
            cwd=frontend_dir,
            capture_output=True,
            text=True,
            encoding='utf-8',
            shell=True
        )

        if result.returncode == 0:
            print_info("前端应用构建成功！")
            return True
        else:
            print_error("前端应用构建失败:")
            print_error(result.stderr)
            return False
    except Exception as e:
        print_error(f"前端构建过程发生错误: {str(e)}")
        return False


def check_frontend_build():
    """检查前端构建文件是否存在"""
    if not os.path.exists(FRONTEND_DIST):
        print_error(f"前端构建目录不存在: {FRONTEND_DIST}")
        return False

    # 检查dist目录是否有文件
    if len(os.listdir(FRONTEND_DIST)) == 0:
        print_error(f"前端构建目录为空: {FRONTEND_DIST}")
        return False

    return True

def prepare_static_directory():
    """准备静态资源目录"""
    # 创建static目录
    if not os.path.exists(STATIC_DIR):
        print_info(f"创建静态资源目录: {STATIC_DIR}")
        os.makedirs(STATIC_DIR)

    # 创建dist子目录
    dist_dir = os.path.join(STATIC_DIR, "dist")
    if os.path.exists(dist_dir):
        print_info(f"清理现有的dist目录: {dist_dir}")
        shutil.rmtree(dist_dir)

    print_info(f"创建dist目录: {dist_dir}")
    os.makedirs(dist_dir)

    return dist_dir

def copy_frontend_files(dist_dir):
    """复制前端构建文件到静态资源目录"""
    print_info(f"复制前端构建文件到: {dist_dir}")

    try:
        # 复制所有文件和子目录
        for item in os.listdir(FRONTEND_DIST):
            src_path = os.path.join(FRONTEND_DIST, item)
            dst_path = os.path.join(dist_dir, item)

            if os.path.isdir(src_path):
                shutil.copytree(src_path, dst_path)
            else:
                shutil.copy2(src_path, dst_path)

        return True
    except Exception as e:
        print_error(f"复制前端文件失败: {str(e)}")
        return False

def prepare_release_directory():
    """准备release目录"""
    print_info(f"准备release目录: {RELEASE_DIR}")

    # 创建release目录
    if not os.path.exists(RELEASE_DIR):
        print_info(f"创建release目录: {RELEASE_DIR}")
        os.makedirs(RELEASE_DIR)
    else:
        # 清理现有内容
        print_info("清理现有的release目录内容...")
        for item in os.listdir(RELEASE_DIR):
            item_path = os.path.join(RELEASE_DIR, item)
            if os.path.isdir(item_path):
                shutil.rmtree(item_path)
            else:
                os.remove(item_path)

    return True


def build_go_application(target_platforms=None, use_cgo=False):
    """构建Go应用，支持多平台交叉编译

    Args:
        target_platforms: 目标平台列表，None则构建当前平台
        use_cgo: 是否启用CGO编译
    """
    print_info("开始构建Go应用...")

    if target_platforms is None:
        target_platforms = [get_current_platform()]

    success_count = 0
    for platform_key in target_platforms:
        if platform_key not in PLATFORMS:
            print_warning(f"未知平台: {platform_key}，跳过")
            continue

        config = PLATFORMS[platform_key]
        print_info(f"构建平台: {platform_key}")

        if build_for_platform(platform_key, config, use_cgo):
            success_count += 1

    print_info(f"构建完成: {success_count}/{len(target_platforms)} 个平台成功")
    return success_count == len(target_platforms)


def get_current_platform():
    """获取当前系统平台标识"""
    system = platform.system().lower()
    machine = platform.machine().lower()

    arch_map = {"x86_64": "amd64", "amd64": "amd64", "arm64": "arm64", "aarch64": "arm64"}
    arch = arch_map.get(machine, machine)

    return f"{system}-{arch}"


def setup_msys2_env(env, msys2_path):
    """设置MSYS2环境变量

    Args:
        env: 环境变量字典
        msys2_path: MSYS2安装路径
    """
    bin_path = os.path.join(msys2_path, "bin")
    current_path = env.get("PATH", "")

    if bin_path not in current_path:
        env["PATH"] = f"{bin_path};{current_path}"

    return env


def setup_cgo_env(env, platform_key):
    """设置CGO编译环境变量

    Args:
        env: 环境变量字典
        platform_key: 目标平台标识
    """
    env["CGO_ENABLED"] = "1"

    if platform_key not in CGO_TOOLCHAINS:
        return env

    toolchain = CGO_TOOLCHAINS[platform_key]
    msys2_path = toolchain.get("msys2_path")

    if msys2_path and os.path.exists(msys2_path):
        setup_msys2_env(env, msys2_path)

    env["CC"] = toolchain["cc"]
    env["CXX"] = toolchain["cxx"]

    return env


def build_for_platform(platform_key, config, use_cgo=False):
    """构建指定平台的可执行文件

    Args:
        platform_key: 平台标识，如 'windows-amd64'
        config: 平台配置字典
        use_cgo: 是否启用CGO编译
    """
    goos = config["goos"]
    goarch = config["goarch"]
    ext = config["ext"]

    platform_dir = os.path.join(RELEASE_DIR, platform_key)
    os.makedirs(platform_dir, exist_ok=True)

    output_exe = os.path.join(platform_dir, f"aiflow{ext}")

    env = os.environ.copy()
    env["GOOS"] = goos
    env["GOARCH"] = goarch

    if use_cgo:
        setup_cgo_env(env, platform_key)
        toolchain = CGO_TOOLCHAINS.get(platform_key, {})
        print_info(f"使用CGO编译: {toolchain.get('description', 'default')}")
    elif goos == "linux":
        env["CGO_ENABLED"] = "1"
        print_warning(f"Linux平台需要CGO和GTK开发库支持")
        print_warning(f"确保已安装: gcc, libgtk-3-dev (Debian/Ubuntu) 或 gtk3-devel (Fedora/RHEL)")
    elif goos == "darwin":
        env["CGO_ENABLED"] = "1"
        print_warning(f"macOS平台需要CGO支持，交叉编译可能受限")
    else:
        env["CGO_ENABLED"] = "0"

    if goos == "windows":
        return build_windows_with_icon(env, output_exe, platform_dir, use_cgo)
    else:
        return build_standard(env, output_exe)


def build_windows_with_icon(env, output_exe, platform_dir, use_cgo=False):
    """构建Windows平台带图标的可执行文件

    Args:
        env: 环境变量字典
        output_exe: 输出文件路径
        platform_dir: 平台输出目录
        use_cgo: 是否启用CGO编译
    """
    try:
        goarch = env.get("GOARCH", "")

        if goarch == "arm64":
            print_warning("rsrc工具不支持ARM64架构，将构建无图标版本")
            return build_standard(env, output_exe)

        if not check_rsrc_tool():
            print_warning("rsrc工具不可用，将构建无图标版本")
            return build_standard(env, output_exe)

        icon_path = os.path.join(GOEND_DIR, "app.ico")
        rsrc_file = os.path.join(API_DIR, "rsrc.syso")

        if not os.path.exists(icon_path):
            print_warning(f"图标文件不存在: {icon_path}，将构建无图标版本")
            return build_standard(env, output_exe)

        print_info(f"生成资源文件，使用图标: {icon_path}")

        result = subprocess.run(
            ["rsrc", "-ico", icon_path, "-o", rsrc_file],
            cwd=GOEND_DIR,
            capture_output=True,
            text=True,
            encoding='utf-8'
        )

        if result.returncode != 0:
            print_warning(f"生成资源文件失败: {result.stderr}，将构建无图标版本")
            return build_standard(env, output_exe)

        success = build_standard(env, output_exe)

        if os.path.exists(rsrc_file):
            os.remove(rsrc_file)
            print_info(f"清理资源文件: {rsrc_file}")

        return success

    except Exception as e:
        print_error(f"Windows构建过程发生错误: {str(e)}")
        return False


def build_standard(env, output_exe):
    """标准构建流程

    Args:
        env: 环境变量字典
        output_exe: 输出文件路径
    """
    try:
        goos = env.get("GOOS", "")
        ldflags = "-s -w"
        if goos == "windows":
            ldflags += " -H windowsgui"

        print_info(f"执行Go构建命令，输出: {output_exe}")
        result = subprocess.run(
            ["go", "build", f"-ldflags={ldflags}", "-o", output_exe, "./cmd/api"],
            cwd=GOEND_DIR,
            capture_output=True,
            text=True,
            encoding='utf-8',
            env=env
        )

        if result.returncode == 0:
            print_info(f"构建成功: {output_exe}")
            return True
        else:
            print_error(f"构建失败: {result.stderr}")
            return False

    except Exception as e:
        print_error(f"构建过程发生错误: {str(e)}")
        return False

def check_rsrc_tool():
    """检查rsrc工具是否可用"""
    try:
        result = subprocess.run(
            ["rsrc"],
            capture_output=True,
            text=True,
            encoding='utf-8'
        )
        # rsrc工具执行成功（即使返回非零退出码），说明工具存在
        return True
    except FileNotFoundError:
        return False

def cleanup():
    """清理临时文件"""
    if os.path.exists(STATIC_DIR):
        print_info(f"清理临时静态资源目录: {STATIC_DIR}")
        # 遍历目录内的所有文件和子目录
        for item in os.listdir(STATIC_DIR):
            item_path = os.path.join(STATIC_DIR, item)
            try:
                if os.path.isdir(item_path):
                    # 使用 onexc 参数忽略删除失败的错误（Windows文件占用问题）
                    shutil.rmtree(item_path, onexc=lambda fn, path, exc: None)
                else:
                    os.remove(item_path)
            except Exception as e:
                print_warning(f"无法删除 {item_path}: {str(e)}")
        # 创建占位文件 placeholder.txt 防止编译器警告
        placeholder_path = os.path.join(STATIC_DIR, "placeholder.txt")
        if not os.path.exists(placeholder_path):
            print_info(f"创建占位文件: {placeholder_path}")
            with open(placeholder_path, "w") as f:
                f.write("这是一个占位文件，防止编译器警告")

def parse_arguments():
    """解析命令行参数"""
    parser = argparse.ArgumentParser(
        description="构建脚本 - 支持多平台交叉编译",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
支持的平台:
  windows-amd64  Windows x64 (支持交叉编译)
  windows-arm64  Windows ARM64 (支持交叉编译，无图标)
  linux-amd64    Linux x64 (需要CGO，建议在Linux上构建)
  linux-arm64    Linux ARM64 (需要CGO，建议在Linux上构建)
  darwin-amd64   macOS Intel (需要CGO，建议在macOS上构建)
  darwin-arm64   macOS Apple Silicon (需要CGO，建议在macOS上构建)

注意事项:
  - Linux/macOS平台需要CGO和GTK依赖，交叉编译受限
  - 建议在对应目标平台上执行构建
  - Windows平台支持从任意系统交叉编译
  - 使用--cgo可启用CGO编译(Windows使用MSYS2 UCRT64 GCC)

示例:
  python builder.py                      # 构建当前平台
  python builder.py --platform all       # 构建所有平台
  python builder.py -p windows-amd64     # 构建Windows x64
  python builder.py -p windows-amd64 -p windows-arm64  # 构建多个Windows平台
  python builder.py --skip-frontend      # 跳过前端构建
  python builder.py -p windows-amd64 --cgo  # 使用CGO编译Windows
        """
    )

    parser.add_argument(
        "-p", "--platform",
        action="append",
        dest="platforms",
        choices=list(PLATFORMS.keys()) + ["all", "current"],
        help="目标平台 (可多次指定，'all'构建所有平台，'current'构建当前平台)"
    )

    parser.add_argument(
        "--skip-frontend",
        action="store_true",
        help="跳过前端构建步骤"
    )

    parser.add_argument(
        "--list-platforms",
        action="store_true",
        help="列出所有支持的平台"
    )

    parser.add_argument(
        "--cgo",
        action="store_true",
        help="启用CGO编译 (Windows使用MSYS2 UCRT64 GCC，Linux/macOS需要对应工具链)"
    )

    return parser.parse_args()


def resolve_target_platforms(requested_platforms):
    """解析目标平台列表

    Args:
        requested_platforms: 用户请求的平台列表

    Returns:
        实际要构建的平台列表
    """
    if not requested_platforms:
        return [get_current_platform()]

    result = []
    for p in requested_platforms:
        if p == "all":
            result.extend(PLATFORMS.keys())
        elif p == "current":
            result.append(get_current_platform())
        else:
            result.append(p)

    return list(dict.fromkeys(result))


def main():
    """主函数"""
    args = parse_arguments()

    if args.list_platforms:
        print("支持的平台:")
        for key, config in PLATFORMS.items():
            print(f"  {key:16} - {config['goos']}/{config['goarch']}")
        return

    print_info("=== 开始构建流程 ===")

    try:
        if args.skip_frontend:
            print_info("跳过前端构建步骤")
            if not check_frontend_build():
                print_error("前端构建文件不存在，请先构建前端或移除--skip-frontend参数")
                sys.exit(1)
        else:
            if not build_frontend():
                sys.exit(1)

        if not check_frontend_build():
            sys.exit(1)

        dist_dir = prepare_static_directory()

        if not copy_frontend_files(dist_dir):
            sys.exit(1)

        if not prepare_release_directory():
            sys.exit(1)

        target_platforms = resolve_target_platforms(args.platforms)
        print_info(f"目标平台: {', '.join(target_platforms)}")

        if args.cgo:
            print_info("启用CGO编译模式")

        if not build_go_application(target_platforms, args.cgo):
            sys.exit(1)

        print_info("=== 构建流程完成 ===")
        print_info(f"发布包位置: {RELEASE_DIR}")
    finally:
        cleanup()


if __name__ == "__main__":
    main()
