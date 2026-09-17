#!/bin/sh
# 一键安装 tcvectordb-gui 到用户目录 (~/.local)
# 用法: ./install.sh 或 sh install.sh
set -eu

APP_NAME="tcvectordb-gui"
SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

BIN_SRC="$SCRIPT_DIR/$APP_NAME"
ICON_SRC="$SCRIPT_DIR/$APP_NAME.png"
DESKTOP_SRC="$SCRIPT_DIR/$APP_NAME.desktop"

[ -f "$BIN_SRC" ] || { echo "错误: 未找到 $BIN_SRC"; exit 1; }
[ -f "$ICON_SRC" ] || { echo "错误: 未找到 $ICON_SRC"; exit 1; }
[ -f "$DESKTOP_SRC" ] || { echo "错误: 未找到 $DESKTOP_SRC"; exit 1; }

BIN_DIR="$HOME/.local/bin"
ICON_DIR="$HOME/.local/share/icons/hicolor/128x128/apps"
DESKTOP_DIR="$HOME/.local/share/applications"

mkdir -p "$BIN_DIR" "$ICON_DIR" "$DESKTOP_DIR"

install -m 755 "$BIN_SRC" "$BIN_DIR/$APP_NAME"
install -m 644 "$ICON_SRC" "$ICON_DIR/$APP_NAME.png"
install -m 644 "$DESKTOP_SRC" "$DESKTOP_DIR/$APP_NAME.desktop"

# 修正 desktop 文件中的 Exec 路径为绝对路径（desktop 规范不支持相对路径）
sed -i "s|^Exec=.*|Exec=$BIN_DIR/$APP_NAME|" "$DESKTOP_DIR/$APP_NAME.desktop"

# 更新桌面数据库与图标缓存（存在才执行）
command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database "$DESKTOP_DIR" || true
command -v gtk-update-icon-cache >/dev/null 2>&1 && gtk-update-icon-cache -f -t "$ICON_DIR" >/dev/null 2>&1 || true

# 确保 ~/.local/bin 在 PATH 中
case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *)
    echo "提示: $BIN_DIR 不在 PATH 中。请将以下内容加入 ~/.profile 或 ~/.bashrc:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    ;;
esac

echo "安装完成: $BIN_DIR/$APP_NAME"
echo "可从应用菜单启动，或运行: $APP_NAME"

# 卸载: rm -f ~/.local/bin/tcvectordb-gui \
#   ~/.local/share/icons/hicolor/128x128/apps/tcvectordb-gui.png \
#   ~/.local/share/applications/tcvectordb-gui.desktop
