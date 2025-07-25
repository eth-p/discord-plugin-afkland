# BetterDiscord Plugin: AFKland
A BetterDiscord plugin to fix AFK detection on Wayland.

## Why?

When Discord thinks you're using the computer, it won't send notifications
to the Discord mobile app. When running under Wayland, Discord always
thinks you're using the computer.

[This is a known issue.](https://support.discord.com/hc/en-us/community/posts/360052371093-Discord-on-Linux-Wayland-has-no-AFK-detection)

## Supported Desktop Environments

 * KDE Plasma 6

If this works on any other environment or if you want to add support for
another environment, pull requests are welcome and appreciated.

## How it Works

A native helper program creates a websocket server on the local machine.
It watches for when the system screensaver is becomes active and informs
any connected websocket clients of the screensaver state.

The BetterDiscord plugin connects to the native helper program and updates
the Discord AFK status based on the screensaver state.

## Installation & Setup

TODO Nix Flake / Home Manager

### Manual

You will need the `go` compiler and `node`/`npm` installed.

 1. Compile the Go program under `src/helper`:

    ```
    go build -o ~/.local/bin/bdplugin-afkland-helper ./src/helper
    ```

 2. Compile the BetterDiscord plugin:

    ```
    npm ci
    npm run build
    ```

 3. Install the systemd service:

    ```
    mkdir -p "$HOME/.local/bin/bdplugin-afkland-helper"
    sed 's#^ExecStart=bdplugin-afkland-helper$#ExecStart='"$HOME/.local/bin/bdplugin-afkland-helper"'#' \
        < assets/bdplugin-afkland-helper.service \
        > ~/.config/systemd/user/bdplugin-afkland-helper.service
    
    systemctl --user enable bdplugin-afkland-helper
    systemctl --user start bdplugin-afkland-helper
    ```

 4. Install the plugin:

    ```
    cp dist/AFKland.plugin.js ~/.var/app/com.discordapp.Discord/config/BetterDiscord/plugins/
    ```

## Alternatives

 * [WayAFK](https://github.com/Colonial-Dev/WayAFK)

Credit to [@Colonial-Dev](https://github.com/Colonial-Dev) for the original idea.
