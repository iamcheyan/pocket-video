import React from 'react';
import { Library } from 'lucide-react';

export function EmptyState({
  title,
  hint,
  action,
  icon,
}: {
  title: string;
  hint?: string;
  action?: { label: string; onClick: () => void };
  icon?: React.ReactNode;
}) {
  return (
    <div className="empty-state">
      <div className="empty-icon">{icon ?? <Library size={20} />}</div>
      <h3>{title}</h3>
      {hint && <p>{hint}</p>}
      {action && (
        <button className="outline" onClick={action.onClick}>
          {action.label}
        </button>
      )}
    </div>
  );
}

export function SyncBadge({ label }: { label: string }) {
  return <span className="sync-badge">{label}</span>;
}
