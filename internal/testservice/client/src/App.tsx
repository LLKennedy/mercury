import './App.css';
import broadcast from './broadcast';
import convert from './convert';
import feed from './feed';

console.log("Client side streaming...")
feed();
// console.log("\nServer side streaming...")
// broadcast();
// console.log("\nDual streaming...")
// convert();

function App() {
  return (
    <div className="App">
      <header className="App-header">

      </header>
    </div>
  );
}

export default App;
