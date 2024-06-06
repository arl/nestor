import './App.css';
import Debugger from './components/Debugger';
import useWS from './ws/hook';

function App() {
  const [, ready] = useWS();

  return (
    <div className="App">
      {!ready && <div>Connecting...</div>}
      <Debugger />
    </div>
  );
}

export default App;
