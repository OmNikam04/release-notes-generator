import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { Provider } from 'react-redux'
import store from './store/store'
import Login from './pages/Login'
import Home from './pages/Home'
import MyBugs from './pages/MyBugs'
import './App.css'

function App() {
  return (
    <Provider store={store}>
      <Router>
        <div className="app">
          <Routes>
            <Route path="/" element={<Login />} />
            <Route path="/home" element={<Home />} />
            <Route path="/mybugs" element={<MyBugs />} />
          </Routes>
        </div>
      </Router>
    </Provider>
  )
}

export default App
