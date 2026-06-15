import React from 'react'

interface SidebarProps {
  activeSection: string
  onSectionChange: (section: string) => void
}

const Sidebar: React.FC<SidebarProps> = ({ activeSection, onSectionChange }) => {
  const sectionLabels: Record<string, string> = {
    transactions: 'Transações',
    balances: 'Saldos',
    wallets: 'Carteiras',
    monitoring: 'Monitoramento',
  }
  const sections = Object.keys(sectionLabels)

  return (
    <aside className="sidebar">
      <ul>
        {sections.map(section => (
          <li key={section}>
            <button
              className={activeSection === section ? 'active' : ''}
              onClick={() => onSectionChange(section)}
            >
              {section.charAt(0).toUpperCase() + section.slice(1)}
            </button>
          </li>
        ))}
      </ul>
    </aside>
  )
}

export default Sidebar