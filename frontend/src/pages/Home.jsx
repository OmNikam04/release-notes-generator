import { useNavigate } from 'react-router-dom';
import './Home.css';

const Home = () => {
  const navigate = useNavigate();

  return (
    <div className="home-page">
      <div className="home-hero">
        <div className="hero-content">
          <div className="quote-section">
            <p className="quote-text">
              "Quality is not an act, it is a habit."
            </p>
            <p className="quote-author">— Aristotle</p>
          </div>

          <h1 className="hero-title">🔥 NoteForge</h1>
          <p className="hero-subtitle">
            Transform your bug fixes into professional release notes with AI-powered automation
          </p>

          <div className="hero-actions">
            <button
              className="btn-primary"
              onClick={() => navigate('/mybugs')}
            >
              Go to My Bugs →
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Home;
