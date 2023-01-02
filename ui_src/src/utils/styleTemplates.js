// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

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
        case 'purple':
            return '#6557FF';
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
