// Configurações globais
const UPDATE_INTERVAL = 2000; // 2 segundos
const MAX_DATA_POINTS = 100;  // Máximo de pontos nos gráficos
const CHART_COLORS = {
    temperature: 'rgb(255, 99, 132)',
    humidity: 'rgb(54, 162, 235)',
    light: 'rgb(255, 205, 86)',
    pressure: 'rgb(75, 192, 192)'
};

// Armazenamento de dados
const sensorData = {
    temperature: [],
    humidity: [],
    light: [],
    pressure: []
};

// Inicialização de gráficos
let tempHumidChart;
let lightPressureChart;
let historyChart;

// Função para resetar a simulação
async function resetSimulation() {
    try {
        const resetButton = document.getElementById('reset-simulation');
        if (resetButton) {
            resetButton.disabled = true;
            resetButton.textContent = 'Resetando...';
        }
        
        const response = await fetch('/api/reset-simulation', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error('Falha ao resetar simulação');
        }
        
        // Limpar os dados armazenados
        Object.keys(sensorData).forEach(key => {
            sensorData[key] = [];
        });
        
        // Atualizar gráficos
        updateCharts();
        
        // Buscar novos dados imediatamente
        await fetchReadings();
        
        // Restaurar botão
        if (resetButton) {
            resetButton.disabled = false;
            resetButton.textContent = 'Resetar Simulação';
        }
        
        console.log('Simulação resetada com sucesso');
    } catch (error) {
        console.error('Erro ao resetar simulação:', error);
        
        const resetButton = document.getElementById('reset-simulation');
        if (resetButton) {
            resetButton.disabled = false;
            resetButton.textContent = 'Resetar Simulação';
        }
    }
}

// Função para inicializar gráficos
function initCharts() {
    // Configuração para o gráfico de temperatura e umidade
    const tempHumidCtx = document.getElementById('temp-humid-chart').getContext('2d');
    tempHumidChart = new Chart(tempHumidCtx, {
        type: 'line',
        data: {
            datasets: [
                {
                    label: 'Temperatura (°C)',
                    borderColor: CHART_COLORS.temperature,
                    backgroundColor: CHART_COLORS.temperature + '20',
                    borderWidth: 2,
                    tension: 0.3,
                    fill: true,
                    data: []
                },
                {
                    label: 'Umidade (%)',
                    borderColor: CHART_COLORS.humidity,
                    backgroundColor: CHART_COLORS.humidity + '20',
                    borderWidth: 2,
                    tension: 0.3,
                    fill: true,
                    data: []
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false
            },
            scales: {
                x: {
                    type: 'time',
                    time: {
                        unit: 'minute',
                        displayFormats: {
                            minute: 'HH:mm:ss'
                        }
                    },
                    title: {
                        display: true,
                        text: 'Hora'
                    }
                },
                y: {
                    title: {
                        display: true,
                        text: 'Valor'
                    }
                }
            },
            plugins: {
                legend: {
                    position: 'top',
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            let label = context.dataset.label || '';
                            if (label) {
                                label += ': ';
                            }
                            if (context.parsed.y !== null) {
                                label += context.parsed.y.toFixed(2);
                            }
                            return label;
                        }
                    }
                }
            }
        }
    });

    // Configuração para o gráfico de luminosidade e pressão
    const lightPressureCtx = document.getElementById('light-pressure-chart').getContext('2d');
    lightPressureChart = new Chart(lightPressureCtx, {
        type: 'line',
        data: {
            datasets: [
                {
                    label: 'Luminosidade (lux)',
                    borderColor: CHART_COLORS.light,
                    backgroundColor: CHART_COLORS.light + '20',
                    borderWidth: 2,
                    tension: 0.3,
                    fill: true,
                    data: [],
                    yAxisID: 'y-light'
                },
                {
                    label: 'Pressão (hPa)',
                    borderColor: CHART_COLORS.pressure,
                    backgroundColor: CHART_COLORS.pressure + '20',
                    borderWidth: 2,
                    tension: 0.3,
                    fill: true,
                    data: [],
                    yAxisID: 'y-pressure'
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false
            },
            scales: {
                x: {
                    type: 'time',
                    time: {
                        unit: 'minute',
                        displayFormats: {
                            minute: 'HH:mm:ss'
                        }
                    },
                    title: {
                        display: true,
                        text: 'Hora'
                    }
                },
                'y-light': {
                    type: 'linear',
                    position: 'left',
                    title: {
                        display: true,
                        text: 'Luminosidade (lux)'
                    },
                    grid: {
                        drawOnChartArea: false
                    }
                },
                'y-pressure': {
                    type: 'linear',
                    position: 'right',
                    title: {
                        display: true,
                        text: 'Pressão (hPa)'
                    }
                }
            },
            plugins: {
                legend: {
                    position: 'top',
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            let label = context.dataset.label || '';
                            if (label) {
                                label += ': ';
                            }
                            if (context.parsed.y !== null) {
                                label += context.parsed.y.toFixed(2);
                            }
                            return label;
                        }
                    }
                }
            }
        }
    });

    // Configuração para o gráfico histórico
    const historyCtx = document.getElementById('history-chart').getContext('2d');
    historyChart = new Chart(historyCtx, {
        type: 'line',
        data: {
            datasets: [
                {
                    label: 'Temperatura (°C)',
                    borderColor: CHART_COLORS.temperature,
                    backgroundColor: 'transparent',
                    borderWidth: 2,
                    tension: 0.3,
                    data: []
                },
                {
                    label: 'Umidade (%)',
                    borderColor: CHART_COLORS.humidity,
                    backgroundColor: 'transparent',
                    borderWidth: 2,
                    tension: 0.3,
                    data: []
                },
                {
                    label: 'Luminosidade (lux)',
                    borderColor: CHART_COLORS.light,
                    backgroundColor: 'transparent',
                    borderWidth: 2,
                    tension: 0.3,
                    data: []
                },
                {
                    label: 'Pressão (hPa)',
                    borderColor: CHART_COLORS.pressure,
                    backgroundColor: 'transparent',
                    borderWidth: 2,
                    tension: 0.3,
                    data: []
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false
            },
            scales: {
                x: {
                    type: 'time',
                    time: {
                        unit: 'minute',
                        displayFormats: {
                            minute: 'HH:mm:ss'
                        }
                    },
                    title: {
                        display: true,
                        text: 'Hora'
                    }
                },
                y: {
                    title: {
                        display: true,
                        text: 'Valor'
                    }
                }
            },
            plugins: {
                legend: {
                    position: 'top',
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            let label = context.dataset.label || '';
                            if (label) {
                                label += ': ';
                            }
                            if (context.parsed.y !== null) {
                                label += context.parsed.y.toFixed(2);
                            }
                            return label;
                        }
                    }
                }
            }
        }
    });
}

// Função para atualizar os valores exibidos nos cards
function updateSensorValues(readings) {
    readings.forEach(reading => {
        const sensorType = reading.sensor_type;
        const sensorId = reading.sensor_id;
        const value = reading.value.toFixed(2);
        
        // Atualizar o valor no card
        const valueElement = document.getElementById(`value-${sensorType}-${sensorId}`);
        if (valueElement) {
            valueElement.textContent = value;
        }
        
        // Armazenar dados para os gráficos
        if (sensorData[sensorType]) {
            sensorData[sensorType].push({
                x: new Date(reading.timestamp),
                y: reading.value
            });
            
            // Limitar o número de pontos
            if (sensorData[sensorType].length > MAX_DATA_POINTS) {
                sensorData[sensorType].shift();
            }
        }
    });
    
    // Atualizar os gráficos
    updateCharts();
}

// Função para atualizar os gráficos
function updateCharts() {
    if (tempHumidChart) {
        tempHumidChart.data.datasets[0].data = sensorData.temperature;
        tempHumidChart.data.datasets[1].data = sensorData.humidity;
        tempHumidChart.update();
    }
    
    if (lightPressureChart) {
        lightPressureChart.data.datasets[0].data = sensorData.light;
        lightPressureChart.data.datasets[1].data = sensorData.pressure;
        lightPressureChart.update();
    }
    
    if (historyChart) {
        historyChart.data.datasets[0].data = sensorData.temperature;
        historyChart.data.datasets[1].data = sensorData.humidity;
        historyChart.data.datasets[2].data = sensorData.light;
        historyChart.data.datasets[3].data = sensorData.pressure;
        historyChart.update();
    }
}

// Função para buscar leituras mais recentes
async function fetchReadings() {
    try {
        const response = await fetch('/api/readings');
        if (!response.ok) {
            throw new Error('Falha ao buscar dados dos sensores');
        }
        
        const readings = await response.json();
        if (readings && readings.length > 0) {
            updateSensorValues(readings);
        }
    } catch (error) {
        console.error('Erro ao buscar leituras:', error);
    }
}

// Função para atualizar status dos serviços
function updateServiceStatus() {
    // Simulação do status dos serviços - em um cenário real isso viria da API
    const mqttStatus = Math.random() > 0.2 ? 'ok' : 'error';
    const opcuaStatus = Math.random() > 0.2 ? 'ok' : 'error';
    const vpnStatus = Math.random() > 0.2 ? 'ok' : 'warning';
    
    // Atualizar indicadores visuais
    document.getElementById('mqtt-status').className = `status-indicator status-${mqttStatus}`;
    document.getElementById('opcua-status').className = `status-indicator status-${opcuaStatus}`;
    document.getElementById('vpn-status').className = `status-indicator status-${vpnStatus}`;
}

// Inicialização
document.addEventListener('DOMContentLoaded', () => {
    // Inicializar gráficos
    initCharts();
    
    // Primeira leitura
    fetchReadings();
    updateServiceStatus();
    
    // Configurar atualização periódica
    setInterval(fetchReadings, UPDATE_INTERVAL);
    setInterval(updateServiceStatus, 5000); // 5 segundos
    
    // Adicionar botão de reset no header
    const container = document.querySelector('.container h2').parentNode;
    if (container) {
        const buttonDiv = document.createElement('div');
        buttonDiv.className = 'text-end mb-3';
        buttonDiv.innerHTML = `
            <button id="reset-simulation" class="btn btn-success">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-arrow-repeat me-1" viewBox="0 0 16 16">
                    <path d="M11.534 7h3.932a.25.25 0 0 1 .192.41l-1.966 2.36a.25.25 0 0 1-.384 0l-1.966-2.36a.25.25 0 0 1 .192-.41zm-11 2h3.932a.25.25 0 0 0 .192-.41L2.692 6.23a.25.25 0 0 0-.384 0L.342 8.59A.25.25 0 0 0 .534 9z"/>
                    <path fill-rule="evenodd" d="M8 3c-1.552 0-2.94.707-3.857 1.818a.5.5 0 1 1-.771-.636A6.002 6.002 0 0 1 13.917 7H12.9A5.002 5.002 0 0 0 8 3zM3.1 9a5.002 5.002 0 0 0 8.757 2.182.5.5 0 1 1 .771.636A6.002 6.002 0 0 1 2.083 9H3.1z"/>
                </svg>
                Resetar Simulação
            </button>
        `;
        container.prepend(buttonDiv);
        
        // Adicionar evento ao botão
        document.getElementById('reset-simulation').addEventListener('click', resetSimulation);
    }
}); 