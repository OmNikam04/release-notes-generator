import './BugCard.css';

const BugCard = ({ bug, onClick }) => {
  const getStatusColor = (status) => {
    const colors = {
      'pending': '#f39c12',
      'dev_approved': '#3498db',
      'mgr_approved': '#9b59b6',
      'approved': '#27ae60',
      'rejected': '#e74c3c'
    };
    return colors[status] || '#95a5a6';
  };

  const getStatusLabel = (status) => {
    const labels = {
      'pending': 'Pending',
      'dev_approved': 'Dev Approved',
      'mgr_approved': 'Mgr Approved',
      'approved': 'Approved',
      'rejected': 'Rejected'
    };
    return labels[status] || status;
  };

  return (
    <div className="bug-card" onClick={onClick}>
      <div className="bug-card-header">
        <div className="bug-header-top">
          <span className="bug-id">{bug.bugsby_id}</span>
          <span
            className="status-badge"
            style={{ backgroundColor: getStatusColor(bug.status) }}
          >
            {getStatusLabel(bug.status)}
          </span>
        </div>
        <h3 className="bug-title">{bug.title}</h3>
      </div>

      <div className="bug-card-body">
        <div className="bug-meta">
          <div className="bug-field">
            <span className="field-label">📦 Component:</span>
            <span className="field-value">{bug.component}</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default BugCard;
