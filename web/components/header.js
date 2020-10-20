function Header({ className, children }) {
  return (
    <header className={className}>
      <nav>
        <a href="/docs">Docs</a>
        <a href="https://github.com/replicate/replicate">GitHub</a>
      </nav>
      <h1 className="tagline">
        <strong>Replicate</strong> Version control for machine&nbsp;learning
      </h1>
      {children}
    </header>
  );
}
export default Header;
