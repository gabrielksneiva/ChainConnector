import React from 'react'
import { useInterests } from '../hooks/useInterests'

const safeJoin = (items?: string[]) => (Array.isArray(items) && items.length > 0 ? items.join(', ') : 'N/A')
const safeJoinTopics = (topics?: string[][]) => (
  Array.isArray(topics) && topics.length > 0 ? topics.map((group) => group.join(', ')).join(' | ') : 'N/A'
)

const InterestList: React.FC = () => {
  const { interests, loading, error } = useInterests()

  if (loading) return <p>Loading interests...</p>
  if (error) return <p>Error: {error}</p>

  return (
    <div className="interest-list">
      <h3>Interesses Registrados</h3>
      {interests.length === 0 ? (
        <p>Nenhum interesse registrado.</p>
      ) : (
        <ul>
          {interests.map((interest, index) => (
            <li key={index}>
              <strong>Enderecos:</strong> {safeJoin(interest.addresses)}<br />
              <strong>Topicos:</strong> {safeJoinTopics(interest.topics)}<br />
              <strong>Hashes:</strong> {safeJoin(interest.tx_hashes)}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

export default InterestList
