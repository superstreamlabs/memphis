import React, { useState, useEffect } from 'react';
import ReactDOM from 'react-dom';
import { Gauge, G2 } from '@ant-design/plots';

const CustomGauge = () => {
    const { registerShape, Util } = G2;

    registerShape('point', 'custom-gauge-indicator2', {
        draw(cfg, container) {
            const { indicator, defaultColor } = cfg.customInfo;
            const { pointer, pin } = indicator;
            const group = container.addGroup();

            const center = this.parsePoint({
                x: 0,
                y: 0
            });
            指针;

            if (pointer) {
                const { startAngle, endAngle } = Util.getAngle(cfg, this.coordinate);
                const radius = this.coordinate.getRadius();
                const midAngle = (startAngle + endAngle) / 2;
                const { x: x1, y: y1 } = Util.polarToCartesian(center.x, center.y, radius / 15, midAngle + 1 / Math.PI);
                const { x: x2, y: y2 } = Util.polarToCartesian(center.x, center.y, radius / 15, midAngle - 1 / Math.PI);
                const { x, y } = Util.polarToCartesian(center.x, center.y, radius * 0.65, midAngle);
                const { x: x0, y: y0 } = Util.polarToCartesian(center.x, center.y, radius * 0.1, midAngle + Math.PI);
                const path = [['M', x0, y0], ['L', x1, y1], ['L', x, y], ['L', x2, y2], ['Z']]; // pointer

                group.addShape('path', {
                    name: 'pointer',
                    attrs: {
                        path,
                        fill: defaultColor,
                        ...pointer.style
                    }
                });
            }

            if (pin) {
                const pinStyle = pin.style || {};
                const { lineWidth = 2, fill = defaultColor, stroke = defaultColor } = pinStyle;
                const r = 6;
                group.addShape('circle', {
                    name: 'pin-outer',
                    attrs: {
                        x: center.x,
                        y: center.y,
                        ...pin.style,
                        fill: 'transparent',
                        r: r * 1.5,
                        lineWidth,
                        stroke: stroke
                    }
                });
                group.addShape('circle', {
                    name: 'pin-inner',
                    attrs: {
                        x: center.x,
                        y: center.y,
                        r,
                        stroke: 'transparent',
                        fill
                    }
                });
            }

            return group;
        }
    });
    const config = {
        percent: 0.78,
        startAngle: Math.PI,
        endAngle: 2 * Math.PI,
        range: {
            color: '#30BF78'
        },
        indicator: {
            shape: 'custom-gauge-indicator',
            pointer: {
                style: {
                    stroke: '#D0D0D0',
                    lineWidth: 1,
                    fill: '#D0D0D0'
                }
            },
            pin: {
                style: {
                    lineWidth: 2,
                    stroke: '#D0D0D0',
                    fill: '#D0D0D0'
                }
            }
        }
    };
    return <Gauge {...config} />;
};

export default CustomGauge;
