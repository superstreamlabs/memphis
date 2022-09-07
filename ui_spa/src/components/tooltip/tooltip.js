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

import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Zoom from '@material-ui/core/Zoom';
import Tooltip from '@material-ui/core/Tooltip';

const TooltipComponent = (props) => {
    const tooltipStyle = makeStyles((theme) => ({
        tooltip: {
            color: props.color === 'white' ? '#2B2E3F' : '#f7f7f7',
            backgroundColor: props.color === 'white' ? '#f7f7f7' : '#2B2E3F',
            fontSize: '12px',
            fontWeight: 400,
            margin: '5px',
            // textAlign: "center",
            boxShadow: 'rgba(0, 0, 0, 0.24) 0px 3px 8px',
            whiteSpace: 'pre-line',
            minWidth: props.minWidth || '60px'
        },
        arrow: {
            color: props.color === 'white' ? '#f7f7f7' : '#2B2E3F'
        }
    }));
    const classes = tooltipStyle();
    const { text } = props;

    return (
        <Tooltip TransitionComponent={Zoom} title={text ? text : ''} classes={classes} arrow>
            {props.children}
        </Tooltip>
    );
};

export default TooltipComponent;
