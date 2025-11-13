import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { login } from '../store/slices/authSlice';
import { authAPI } from '../services/api';
import './Login.css';

const Login = () => {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('developer');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const dispatch = useDispatch();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    if (!email || !role) {
      setError('Please fill in all fields');
      setLoading(false);
      return;
    }

    try {
      console.log('[LOGIN] Starting login process with email:', email, 'role:', role);

      // Call backend API
      const loginData = await authAPI.login(email, role);

      console.log('[LOGIN] Login successful! User data from DB:', loginData.user);
      console.log('[LOGIN] User ID:', loginData.user.id);
      console.log('[LOGIN] User Email:', loginData.user.email);
      console.log('[LOGIN] User Role:', loginData.user.role);
      console.log('[LOGIN] Created At:', loginData.user.created_at);
      console.log('[LOGIN] Updated At:', loginData.user.updated_at);

      // Store tokens
      localStorage.setItem('authToken', loginData.token);
      localStorage.setItem('refreshToken', loginData.refresh_token);
      localStorage.setItem('userEmail', loginData.user.email);
      localStorage.setItem('userRole', loginData.user.role);
      localStorage.setItem('userId', loginData.user.id);

      console.log('[LOGIN] Tokens and user data stored in localStorage');

      // Dispatch to Redux
      dispatch(login({
        email: loginData.user.email,
        role: loginData.user.role,
        id: loginData.user.id,
      }));

      console.log('[LOGIN] Redux state updated');
      console.log('[LOGIN] Redirecting to /home');

      navigate('/home');
    } catch (err) {
      console.error('[LOGIN] Login failed:', err.message);
      setError(err.message || 'Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-box">
        <div className="login-header">
          <h1 className="login-title">Release Notes Generator</h1>
          <p className="login-subtitle">Sign in to continue</p>
        </div>

        <form className="login-form" onSubmit={handleSubmit}>
          {error && <div className="error-message">{error}</div>}

          <div className="form-group">
            <label htmlFor="email" className="form-label">Email</label>
            <input
              type="email"
              id="email"
              className="form-input"
              placeholder="Enter your email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={loading}
              required
            />
          </div>

          <div className="form-group">
            <label className="form-label">Role</label>
            <div className="role-options">
              <label className="role-option">
                <input
                  type="radio"
                  name="role"
                  value="developer"
                  checked={role === 'developer'}
                  onChange={(e) => setRole(e.target.value)}
                  disabled={loading}
                />
                <span className="role-label">Developer</span>
              </label>
              <label className="role-option">
                <input
                  type="radio"
                  name="role"
                  value="manager"
                  checked={role === 'manager'}
                  onChange={(e) => setRole(e.target.value)}
                  disabled={loading}
                />
                <span className="role-label">Manager</span>
              </label>
            </div>
          </div>

          <button type="submit" className="login-button" disabled={loading}>
            {loading ? 'Signing In...' : 'Sign In'}
          </button>
        </form>

        <div className="login-footer">
          <p className="footer-text">Arista Networks © 2024</p>
        </div>
      </div>
    </div>
  );
};

export default Login;

