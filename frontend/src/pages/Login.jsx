import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { login } from '../store/slices/authSlice';
import './Login.css';

const Login = () => {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('developer');
  const navigate = useNavigate();
  const dispatch = useDispatch();

  const handleSubmit = (e) => {
    e.preventDefault();
    // TODO: Add actual authentication logic
    if (email && role) {
      // Dispatch login action to Redux store
      dispatch(login({ email, role }));

      // Also store in localStorage for persistence
      localStorage.setItem('userEmail', email);
      localStorage.setItem('userRole', role);

      navigate('/home');
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
          <div className="form-group">
            <label htmlFor="email" className="form-label">Email</label>
            <input
              type="email"
              id="email"
              className="form-input"
              placeholder="Enter your email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
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
                />
                <span className="role-label">Manager</span>
              </label>
            </div>
          </div>

          <button type="submit" className="login-button">
            Sign In
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

