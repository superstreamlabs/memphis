// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

export function getBorderRadius(radiusType) {
    switch (radiusType) {
        case 'circle':
            return '50px';
        case 'square':
            return '0px';
        case 'semi-round':
            return '5px';
        default:
            return '0px';
    }
}

export function getBorderColor(borderColorType) {
    switch (borderColorType) {
        case 'none':
            return 'transparent';
        case 'gray':
            return '#d8d8d8';
        case 'gray-light':
            return '#E9E9E9';
        case 'purple':
            return '#6557FF';
        case 'navy':
            return '#1D1D1D';
        case 'search-input':
            return '#5A4FE5';
        case 'white':
            return '#ffffff';
        default:
            return borderColorType;
    }
}

export function getFontColor(colorType) {
    switch (colorType) {
        case 'none':
            return 'transparent';
        case 'black':
            return '#1D1D1D';
        case 'purple':
            return '#6557FF';
        case 'navy':
            return '#1D1D1D';
        case 'gray':
            return '#A9A9A9';
        case 'gray-dark':
            return 'rgba(74, 73, 92, 0.8)';
        case 'white':
            return '#ffffff';
        case 'red':
            return '#FF4838';
        default:
            return '#6557FF';
    }
}

export function getBackgroundColor(backgroundColor) {
    switch (backgroundColor) {
        case 'green':
            return '#27AE60';
        case 'purple':
            return '#6557FF';
        case 'purple-light':
            return '#D0CCFF';
        case 'white':
            return '#FFFFFF';
        case 'orange':
            return '#FFC633';
        case 'red':
            return '#E54F4F';
        case 'navy':
            return '#1D1D1D';
        case 'turquoise':
            return '#5CA6A0';
        case 'black':
            return '#18171E';
        case 'gray':
            return '#A9A9A9';
        case 'gray-light':
            return '#E9E9E9';
        case 'gray-dark':
            return '#EBEDF0';
        case 'disabled':
            return '#F5F5F5';
        case 'none':
            return 'transparent';
        default:
            return '#F0F1F7';
    }
}

export function getBoxShadows(boxShadowsType) {
    switch (boxShadowsType) {
        case 'none':
            return 'none';
        case 'gray':
            return '0px 0px 2px 0px rgba(0,0,0,0.5)';
        case 'gray2':
            return '0px 1px 2px 0px rgba(0,0,0,0.5)';
        case 'float':
            return '0px 1px 2px 0px rgba(0,0,0,0.21)';
        case 'search-input':
            return '0px 1px 2px 0px rgba(90, 79, 229, 1)';
        default:
            return 'none';
    }
}
