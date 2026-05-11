import React, { useState, useEffect } from 'react';
import { Scanner } from '@yudiel/react-qr-scanner';

function App() {
  const [clases, setClases] = useState([]);
  const [seleccion, setSeleccion] = useState(null);
  const [mensaje, setMensaje] = useState("");
  const [escaneando, setEscaneando] = useState(true);

  // ACTUALIZADO: Usando la nueva IP Elástica 54.173.32.242
  const API_BASE = "http://54.173.32.242:3000";

  useEffect(() => {
    fetch(`${API_BASE}/clases`)
      .then(res => res.json())
      .then(data => setClases(data))
      .catch(err => console.error("Error cargando clases:", err));
  }, []);

  const handleScan = (result) => {
    if (result && result[0]?.rawValue && escaneando) {
      setEscaneando(false);
      
      fetch(`${API_BASE}/api/asistencia`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          alumno_id: result[0].rawValue, 
          clase_id: seleccion.id 
        })
      })
      .then(res => res.json())
      .then(data => {
        setMensaje(data.mensaje);
        setTimeout(() => {
          setMensaje("");
          setEscaneando(true);
        }, 3000);
      })
      .catch(err => {
        console.error("Error en asistencia:", err);
        setEscaneando(true);
      });
    }
  };

  const enviarReporte = () => {
    if(window.confirm("¿Enviar reporte del grupo 6N LIDTS?")) {
      fetch(`${API_BASE}/api/cerrar-clase`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ clase_id: seleccion.id })
      })
      .then(res => res.json())
      .then(data => alert(data.mensaje))
      .catch(err => alert("Error al enviar reporte"));
    }
  };

  return (
    <div style={{ display: 'flex', minHeight: '100vh', fontFamily: 'Arial', backgroundColor: '#121212', color: 'white' }}>
      <div style={{ width: '250px', backgroundColor: '#003b70', padding: '20px' }}>
        <h2>SIAE</h2>
        <p>UNACH - LIDTS</p>
        <p style={{fontSize: '0.8em', color: '#ccc'}}>Grupo: 6N LIDTS</p>
      </div>

      <div style={{ flex: 1, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
        <div style={{ backgroundColor: 'white', color: '#333', padding: '30px', borderRadius: '15px', width: '450px', textAlign: 'center' }}>
          {!seleccion ? (
            <>
              <h3>Seleccione Materia</h3>
              {clases && clases.length > 0 ? clases.map(c => (
                <button 
                  key={c.id} 
                  onClick={() => setSeleccion(c)} 
                  style={{ width: '100%', padding: '12px', margin: '5px 0', cursor: 'pointer', borderRadius: '8px', border: '2px solid #003b70', fontWeight: 'bold', color: '#003b70', background: 'white' }}
                >
                  {c.nombre}
                </button>
              )) : <p>Cargando materias...</p>}
            </>
          ) : (
            <>
              <h3 style={{color: '#003b70'}}>{seleccion.nombre}</h3>
              <div style={{opacity: escaneando ? 1 : 0.5}}>
                <Scanner onScan={handleScan} />
              </div>
              
              {mensaje && (
                <div style={{ margin: '15px 0', padding: '15px', backgroundColor: mensaje.includes("¡Bienvenido") ? '#d4edda' : '#fff3cd', color: '#155724', borderRadius: '8px', fontWeight: 'bold' }}>
                  {mensaje}
                </div>
              )}

              <div style={{marginTop: '20px', display: 'flex', flexDirection: 'column', gap: '10px'}}>
                <button 
                  onClick={enviarReporte} 
                  style={{ width: '100%', padding: '15px', backgroundColor: '#28a745', color: 'white', border: 'none', borderRadius: '8px', fontWeight: 'bold', cursor: 'pointer' }}
                >
                  Finalizar Clase y Enviar Reporte
                </button>
                <button 
                  onClick={() => {setSeleccion(null); setMensaje(""); setEscaneando(true);}} 
                  style={{ color: '#dc3545', cursor: 'pointer', background: 'none', border: 'none' }}
                >
                  Volver / Cambiar Materia
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

export default App;