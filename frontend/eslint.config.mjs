import { dirname } from "path";
import { fileURLToPath } from "url";
import { FlatCompat } from "@eslint/eslintrc";
import stylistic from "@stylistic/eslint-plugin";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const eslintConfig = [
  ...compat.extends("next/core-web-vitals", "next/typescript"),
  stylistic.configs.customize({
    // 基本的な設定
    quotes: "double", // ダブルクォート
    semi: true, // セミコロン必須
    indent: 2, // 2スペースインデント
    jsx: true, // JSX対応
    maxLen: 100,

    // より詳細な設定
    arrowParens: "always", // アロー関数の括弧を常に使用
    braceStyle: "1tbs", // One True Brace Style
    // commaDangle: "es5", // ES5準拠のカンマ
    quoteProps: "as-needed", // 必要な場合のみオブジェクトプロパティをクォート
  }),
];

export default eslintConfig;
