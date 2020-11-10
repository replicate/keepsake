import Link from "next/link";

function Header({ className, children }) {
  return (
    <header className={className}>
      <nav>
        <Link href="/docs">
          <a>Docs</a>
        </Link>
        <a href="https://github.com/replicate/replicate">GitHub</a>
      </nav>
      <h1 className="tagline">
        <Link href="/">
          <a>
            <strong>Replicate</strong> Version control for machine&nbsp;learning
          </a>
        </Link>
      </h1>
      {children}
    </header>
  );
}
export default Header;
