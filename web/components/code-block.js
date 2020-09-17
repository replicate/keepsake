import Highlight, { defaultProps } from "prism-react-renderer";
import Prism from "prism-react-renderer/prism";

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

function CodeBlock({ children, className, ...props }) {
  if (className) {
    const language = className.replace(/language-/, "");
    if (!props.language && language) {
      props.language = language;
    }
  }
  return (
    <div className="codeblock">
      <Highlight
        {...defaultProps}
        code={children.trim()}
        theme={null}
        {...props}
      >
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
    </div>
  );
}
export default CodeBlock;
