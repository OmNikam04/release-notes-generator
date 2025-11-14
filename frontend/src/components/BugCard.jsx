import './BugCard.css';

const BugCard = ({ bug, onClick }) => {
  // Handle both old format (bug.bugsby_id) and new format (bug.bug.bugsby_id)
  const bugData = bug.bug || bug;
  const releaseNoteContent = bug.content || bug.generated_note;

  const getStatusColor = (status) => {
    const colors = {
      'pending': '#f39c12',
      'ai_generated': '#f39c12',
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
      'ai_generated': 'AI Generated',
      'dev_approved': 'Dev Approved',
      'mgr_approved': 'Mgr Approved',
      'approved': 'Approved',
      'rejected': 'Rejected'
    };
    return labels[status] || status;
  };

  const getFirstLineOfNote = (note) => {
    if (!note) return 'No release note generated yet';
    const firstLine = note.split('\n')[0];
    return firstLine.length > 80 ? firstLine.substring(0, 80) + '...' : firstLine;
  };

  return (
    <div className="bug-card" onClick={onClick}>
      {/* Desktop View - Full Details */}
      <div className="bug-card-desktop">
        <div className="bug-card-header">
          <div className="bug-header-top">
            <span className="bug-id">{bugData.bugsby_id}</span>
            <span
              className="status-badge"
              style={{ backgroundColor: getStatusColor(bug.status) }}
            >
              {getStatusLabel(bug.status)}
            </span>
          </div>
          <h3 className="bug-title">{bugData.title}</h3>
        </div>

        <div className="bug-card-body">
          <div className="bug-meta">
            <div className="bug-field">
              <span className="field-label">📝 Release Note:</span>
              <span className="field-value note-preview">{getFirstLineOfNote(releaseNoteContent)}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Mobile View - Compact */}
      <div className="bug-card-mobile">
        <div className="mobile-header">
          <span className="bug-id-mobile">{bugData.bugsby_id}</span>
          <span
            className="status-badge-mobile"
            style={{ backgroundColor: getStatusColor(bug.status) }}
          >
            {getStatusLabel(bug.status)}
          </span>
        </div>
        <h3 className="bug-title-mobile">{bugData.title}</h3>
      </div>
    </div>
  );
};

export default BugCard;
