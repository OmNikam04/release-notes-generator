import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { Provider } from 'react-redux'
import store from './store/store'
import Login from './pages/Login'
import MyBugs from './pages/MyBugs'
import ReleaseAdmin from './pages/ReleaseAdmin'
import './App.css'

function App() {
  return (
    <Provider store={store}>
      <Router>
        <div className="app">
          <Routes>
            <Route path="/" element={<Login />} />
            <Route path="/bugs" element={<MyBugs />} />
            <Route path="/releaseadmin" element={<ReleaseAdmin />} />
          </Routes>
        </div>
      </Router>
    </Provider>
  )
}

export default App
