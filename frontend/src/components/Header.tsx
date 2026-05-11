import React from 'react'

interface HeaderProps {
  title: string
}

const Header: React.FC<HeaderProps> = ({ title }) => {
  return (
    <header className="header">
      <h1>{title}</h1>
      <nav>
        {/* Add navigation items if needed */}
      </nav>
    </header>
  )
}

export default Header