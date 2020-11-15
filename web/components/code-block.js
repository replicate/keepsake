import React, { useState } from "react";
import Highlight, { defaultProps } from "prism-react-renderer";
import Prism from "prism-react-renderer/prism";
import copy from "copy-text-to-clipboard";

(typeof global !== "undefined" ? global : window).Prism = Prism;

require("prismjs/components/prism-shell-session");

// https://github.com/FormidableLabs/prism-react-renderer/issues/26
let highlightStart = false;
const highlightClassName = "line-highlight";

const highlightLine = (lineArray, lineProps) => {
  let shouldExclude = false;

  lineArray.forEach((line, i) => {
    const content = line.content;

    // Highlight lines with "# highlight-line"
    if (content.replace(/\s/g, "").includes("#highlight-line")) {
      lineProps.className = `${lineProps.className} ${highlightClassName}`;
      line.content = content
        .replace("# highlight-line", "")
        .replace("#highlight-line", "");
    }

    // Stop highlighting
    if (!!highlightStart && content.replace(/\s/g, "") === "#highlight-end") {
      highlightStart = false;
      shouldExclude = true;
    }

    // Start highlighting after "#highlight-start"
    if (content.replace(/\s/g, "") === "#highlight-start") {
      highlightStart = true;
      shouldExclude = true;
    }
  });

  // Highlight lines between #highlight-start & #highlight-end
  if (!!highlightStart) {
    lineProps.className = `${lineProps.className} ${highlightClassName}`;
  }

  return shouldExclude;
};

function CodeBlock({ children, className = "", copyButton = true, ...props }) {
  const [showCopied, setShowCopied] = useState(false);
  const code = children.trim();

  if (props.language == "shell-session") {
    copyButton = false;
  }

  return (
    <div className={`codeblock ${className}`}>
      <Highlight {...defaultProps} code={code} theme={null} {...props}>
        {({ className, style, tokens, getLineProps, getTokenProps }) => (
          <pre className={className} style={style}>
            {tokens.map((line, i) => {
              const lineProps = getLineProps({ line, key: i });
              const shouldExclude = highlightLine(line, lineProps);
              return !shouldExclude ? (
                <div {...lineProps}>
                  {line.map((token, key) => (
                    <span {...getTokenProps({ token, key })} />
                  ))}
                </div>
              ) : null;
            })}
          </pre>
        )}
      </Highlight>
      {copyButton ? (
        <button
          type="button"
          className="copy-button"
          aria-label="Copy code to clipboard"
          onClick={() => {
            copy(code);
            setShowCopied(true);
            setTimeout(() => setShowCopied(false), 2000);
          }}
        >
          {showCopied ? "Copied" : "Copy"}
        </button>
      ) : (
        ""
      )}
    </div>
  );
}
export default CodeBlock;
