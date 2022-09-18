// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
        case 'navy':
            return '#1D1D1D';
        default:
            return 'transparent';
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
        case 'white':
            return '#f7f7f7';
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
            return '#CD5C5C';
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
        case 'login-input':
            return '0px 1px 2px 0px rgba(0,0,0,0.21)';
    }
}
