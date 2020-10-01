function Header({ className, children }) {
  return (
    <header className={className}>
      <h1 className="tagline">
        <strong>Replicate</strong> Version control for machine&nbsp;learning
      </h1>
      {children}
    </header>
  );
}
export default Header;
