function Header({ className, children }) {
  return (
    <header className={className}>
      <nav>
        <h1 className="tagline">
          <strong>Replicate</strong> Version control for machine learning
        </h1>
        <a href="/docs">Docs</a>
        <a href="https://github.com/replicate/replicate">GitHub</a>
      </nav>
      {children}
    </header>
  );
}
export default Header;
