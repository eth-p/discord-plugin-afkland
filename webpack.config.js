import webpack from "webpack";
import path from "node:path";
import { readFileSync } from "node:fs";

const generatePluginMeta = () => {
	const packageInfo = readFileSync("./package.json", "utf-8");
	const packageJson = JSON.parse(packageInfo);

	const pluginMeta = {
		...packageJson.betterdiscordPlugin,
		version: packageJson.version,
		author: packageJson.author,
		description: packageJson.description,
	};

	const pluginMetaLines = Object.entries(pluginMeta)
		.map(([name, value]) => ` * @${name} ${value}`)
		.join("\n");

	return `/**\n${pluginMetaLines}\n */`;
};

export default {
	mode: "development",
	target: "node",
	devtool: false,
	entry: "./src/plugin/index.ts",
	output: {
		filename: "AFKland.plugin.js",
		path: path.join(import.meta.dirname, "dist"),
		libraryTarget: "commonjs2",
		libraryExport: "default",
		compareBeforeEmit: false
	},

	resolve: {
		extensions: [".js", ".ts"],
	},

	module: {
		rules: [{ test: /\.ts$/, use: "ts-loader"}],
	},

	plugins: [
		new webpack.BannerPlugin({ raw: true, banner: generatePluginMeta }),
	],
};
