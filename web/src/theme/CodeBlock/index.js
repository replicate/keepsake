import React from "react";
import styles from "./styles.module.scss";
// This might change https://github.com/facebook/docusaurus/issues/3043
import OriginalCodeBlock from "@theme-original/CodeBlock";

export default (props) => (
  <div className={styles.codeBlock}>
    <OriginalCodeBlock {...props} />
  </div>
);
