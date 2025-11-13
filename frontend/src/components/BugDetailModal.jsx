import { useState } from 'react';
import './BugDetailModal.css';

const BugDetailModal = ({ bug, onClose }) => {
  if (!bug) return null;

  const [releaseNote, setReleaseNote] = useState("Fixed critical authentication bug that prevented users from logging in with valid credentials.");

  const alternatives = [
    "Resolved authentication service bug preventing user login with valid credentials.",
    "Fixed critical login issue in authentication service affecting user access.",
    "Authentication bug fix: Users can now successfully log in with their credentials."
  ];

  const handleUseAlternative = (alternative) => {
    setReleaseNote(alternative);
  };

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
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <div className="modal-title-section">
            <h2 className="modal-title">{bug.title}</h2>
            <div className="modal-badges">
              <span className="status-label-text">Status:</span>
              <span
                className="modal-badge status-badge"
                style={{ backgroundColor: getStatusColor(bug.status) }}
              >
                {getStatusLabel(bug.status)}
              </span>
            </div>
          </div>
          <button className="modal-close-btn" onClick={onClose}>
            ✕
          </button>
        </div>

        <div className="modal-body">
          <div className="modal-main-content">
          {/* Bug Details - FIRST */}
          <div className="modal-details-section">
            <div className="modal-section">
              <h3 className="section-title">
                <span className="title-icon">🐛</span>
                Bug Details
              </h3>

              <div className="detail-description">
                <span className="detail-label">Description</span>
                <div className="detail-value description-scroll">
                  {bug.description}
                </div>
              </div>

              <div className="details-grid">
                <div className="detail-item">
                  <span className="detail-label">Bug ID</span>
                  <span className="detail-value">{bug.bugsby_id}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Severity</span>
                  <span className="detail-value severity-badge severity-{bug.severity?.toLowerCase()}">
                    {bug.severity}
                  </span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Priority</span>
                  <span className="detail-value">{bug.priority}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Component</span>
                  <span className="detail-value">{bug.component}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Bug Type</span>
                  <span className="detail-value">{bug.bug_type}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Release</span>
                  <span className="detail-value">{bug.release}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Assigned To</span>
                  <span className="detail-value">{bug.assigned_to}</span>
                </div>
                {bug.cve_number && (
                  <div className="detail-item">
                    <span className="detail-label">CVE Number</span>
                    <span className="detail-value cve-badge">{bug.cve_number}</span>
                  </div>
                )}
                <div className="detail-item">
                  <span className="detail-label">Created At</span>
                  <span className="detail-value">{new Date(bug.created_at).toLocaleString()}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Last Synced</span>
                  <span className="detail-value">{new Date(bug.last_synced_at).toLocaleString()}</span>
                </div>
              </div>
            </div>
          </div>

          {/* Generated Release Note - SECOND */}
          <div className="modal-response-section">
            <div className="modal-section">
              <h3 className="section-title">
                <span className="title-icon">📝</span>
                Generated Release Note
              </h3>
              <textarea
                className="response-editor"
                placeholder="The AI-generated release note will appear here. You can edit it as needed..."
                rows="4"
                value={releaseNote}
                onChange={(e) => setReleaseNote(e.target.value)}
              />

              {/* Alternative Suggestions */}
              <div className="alternative-suggestions">
                <h4 className="suggestions-title">💡 Alternative Versions</h4>
                <div className="suggestions-list">
                  {alternatives.map((alt, index) => (
                    <div key={index} className="suggestion-item">
                      <div className="suggestion-content">{alt}</div>
                      <button
                        className="btn-use-suggestion"
                        onClick={() => handleUseAlternative(alt)}
                      >
                        Use This
                      </button>
                    </div>
                  ))}
                </div>
              </div>

              {/* Feedback for Regeneration - AFTER alternatives */}
              <div className="feedback-section">
                <h4 className="feedback-title">💬 Feedback for Regeneration</h4>
                <p className="feedback-hint">
                  Not satisfied? Add feedback below and regenerate to improve the release note.
                </p>
                <textarea
                  className="feedback-input"
                  placeholder="E.g., 'Make it more technical', 'Add more details about the fix', 'Simplify the language'..."
                  rows="3"
                />
                <button className="btn-regenerate">
                  🔄 Regenerate Release Note
                </button>
              </div>
            </div>
          </div>
          </div>

          {/* Approval Workflow - RIGHT SIDE */}
          <div className="modal-approval-section">
            <div className="approval-card">
              <h3 className="approval-title">📋 Release Note Status</h3>
              <div className="approval-status">
                <span className="status-label">Current Status</span>
                <span className="status-value">Raw AI generate Draft</span>
              </div>
              <div className="approval-actions">
                <button className="btn-approval-primary">
                  ✓ Send to Approval
                </button>
              </div>
              <div className="approval-info">
                <p className="info-text">
                  <strong>Reviewer:</strong> Sarah Wilson
                </p>
                <p className="info-text">
                  <strong>Role:</strong> Manager
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="modal-footer">
          <button className="btn-modal-secondary" onClick={onClose}>
            Close
          </button>
        </div>
      </div>
    </div>
  );
};

export default BugDetailModal;
