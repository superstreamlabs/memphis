import React, { useEffect, useRef } from 'react';
import Chart from 'chart.js';

const HalfDoughnutChart = () => {
    const chartRef = useRef(null);

    useEffect(() => {
        const data = {
            // labels: ['Red', 'Blue', 'Yellow'],
            datasets: [
                {
                    label: 'My First Dataset',
                    data: [300, 50],
                    backgroundColor: ['#6557FF', 'rgba(101, 87, 255, 0.15)']
                }
            ]
        };

        const config = {
            type: 'doughnut',
            data: data,
            cutout: '90%',
            options: {
                responsive: true,
                maintainAspectRatio: false,
                rotation: Math.PI, // Rotate by 90 degrees in radians
                circumference: Math.PI,
                cutout: '100px',
                borderWidth: 0,
                plugins: {
                    title: false,
                    subtitle: false,
                    legend: false,
                    tooltips: false
                },
                tooltips: false,
                elements: {
                    arc: {
                        borderWidth: 1,
                        borderRadius: function (context) {
                            const index = context.dataIndex;
                            return index === 0 ? 1 : 0; // Add a border radius to the first color segment
                        }
                    }
                }
            }
        };

        const ctx = chartRef.current.getContext('2d');
        new Chart(ctx, config);
    }, []);

    return (
        <div className="halfdoughnut">
            <canvas id="myChart" ref={chartRef}></canvas>
        </div>
    );
};

export default HalfDoughnutChart;
