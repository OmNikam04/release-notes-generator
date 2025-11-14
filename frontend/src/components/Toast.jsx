import { useEffect } from 'react';
import './Toast.css';

/**
 * Toast Component - Reusable notification component
 * 
 * @param {Object} props
 * @param {string} props.message - The message to display
 * @param {string} props.type - Type of toast: 'success', 'error', 'info', 'warning'
 * @param {boolean} props.show - Whether to show the toast
 * @param {function} props.onClose - Callback when toast is closed
 * @param {number} props.duration - Auto-hide duration in ms (default: 5000, 0 = no auto-hide)
 * @param {string} props.position - Position: 'top-right', 'top-left', 'bottom-right', 'bottom-left' (default: 'bottom-right')
 */
const Toast = ({ 
  message, 
  type = 'info', 
  show = false, 
  onClose, 
  duration = 5000,
  position = 'bottom-right'
}) => {
  useEffect(() => {
    if (show && duration > 0) {
      const timer = setTimeout(() => {
        onClose?.();
      }, duration);

      return () => clearTimeout(timer);
    }
  }, [show, duration, onClose]);

  if (!show) return null;

  const getIcon = () => {
    switch (type) {
      case 'success':
        return '✅';
      case 'error':
        return '⚠️';
      case 'warning':
        return '⚡';
      case 'info':
      default:
        return 'ℹ️';
    }
  };

  return (
    <div className={`toast toast-${type} toast-${position}`}>
      <div className="toast-content">
        <span className="toast-icon">{getIcon()}</span>
        <span className="toast-message">{message}</span>
      </div>
      {onClose && (
        <button
          className="toast-close-btn"
          onClick={onClose}
          title="Close"
        >
          ✕
        </button>
      )}
    </div>
  );
};

export default Toast;

