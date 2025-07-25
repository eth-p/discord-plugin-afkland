const Dispatcher = BdApi.Webpack.getModule(m => m.dispatch && m.subscribe) as DiscordDispatcher;
const Patcher = BdApi.Patcher;
const UI = BdApi.UI;

interface PluginMeta {
    name: string;
    version: string;
}

interface DiscordDispatcher {
    dispatch(event: { type: string } & any): void
}

export default class AFKland {
    private meta: PluginMeta;
    private afkStatus: boolean;
    private debounceTimeout: number | null;
    private socket: WebSocket | null;

    constructor(meta: PluginMeta) {
        this.meta = meta;
        this.afkStatus = true;
        this.debounceTimeout = null;
        this.socket = null;
    }

    start() {
        this.connectToNativeHelper()

        // Patch the dispatcher to replace the broken Discord-detected
        // AFK status with what our native helper says the AFK status is.
        BdApi.Patcher.before(this.meta.name, Dispatcher, "dispatch", (_, args) => {
            const event = args[0] as any;
            if (event.type === "AFK") {
                event.afk = this.afkStatus;
            }
        });
    }

    connectToNativeHelper() {
        let retryPending = false;
        const { name } = this.meta;
        let socket!: WebSocket;
        
        try {
            socket = new WebSocket("ws://127.0.0.1:16738/afkland-native-helper");
        } catch (err) {
            this.notifyNativeHelperConnectionError(err as Error);
            setTimeout(() => this.connectToNativeHelper, 30000);
            return 
        }

        socket.addEventListener("open", () => {
            BdApi.Logger.info(name, "Connected to native helper.");
        });

        socket.addEventListener("message", (event) => {
            const isAFK = JSON.parse(event.data)
            if ((typeof isAFK) !== "boolean") {
                BdApi.Logger.error(name, "Native helper sent malformed message.", isAFK)
            }

            this.afkStatus = isAFK;
            this.dispatchAFKStatus();
        });

        socket.addEventListener("error", () => {
            if (retryPending) {
                return
            }

            retryPending = true;
            this.notifyNativeHelperConnectionError();
            setTimeout(() => this.connectToNativeHelper, 30000);
        })

        this.socket = socket;
    }

    notifyNativeHelperConnectionError(err?: Error) {
        BdApi.Logger.error(this.meta.name, "Connection error for native helper.", err);
        UI.showToast(`Cannot detect Wayland AFK status.`, {
            type: "warn"
        })
    }

    dispatchAFKStatus() {
        if (this.debounceTimeout != null) {
            clearTimeout(this.debounceTimeout)
            this.debounceTimeout = null;
        }

        this.debounceTimeout = setTimeout(() => {
            Dispatcher.dispatch({
                type: "AFK",
                afk: this.afkStatus,
            })
            this.debounceTimeout = null;
        }, 2500)
    }

    stop() {
        Patcher.unpatchAll(this.meta.name);
        if (this.socket != null) {
            this.socket.close(1000, "plugin unloaded");
        }
    }

};
