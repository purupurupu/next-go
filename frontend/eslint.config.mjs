import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTypescript from "eslint-config-next/typescript";
import stylistic from "@stylistic/eslint-plugin";

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTypescript,
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
    quoteProps: "as-needed", // 必要な場合のみオブジェクトプロパティをクォート
  }),
  globalIgnores([
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
  ]),
]);

export default eslintConfig;
