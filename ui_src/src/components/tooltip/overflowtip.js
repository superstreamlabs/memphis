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

import React, { useRef, useEffect, useState } from 'react';
import Tooltip from '@material-ui/core/Tooltip';
import { makeStyles } from '@material-ui/core/styles';

const OverflowTip = (props) => {
    const tooltipStyle = makeStyles((theme) => ({
        tooltip: {
            color: props.color === 'white' ? '#2B2E3F' : '#f7f7f7',
            backgroundColor: props.color === 'white' ? '#f7f7f7' : '#2B2E3F',
            fontSize: '14px',
            fontWeight: 800,
            margin: '5px',
            fontFamily: 'Inter',
            boxShadow: 'rgba(0, 0, 0, 0.24) 0px 3px 8px',
            whiteSpace: 'pre-line',
            textAlign: props.center ? 'center' : 'start'
        },
        arrow: {
            color: props.color === 'white' ? '#f7f7f7' : '#2B2E3F'
        }
    }));
    const classes = tooltipStyle();
    // Create Ref
    const textElementRef = useRef();

    const compareSize = () => {
        const compare = textElementRef.current.scrollWidth > textElementRef.current.clientWidth;
        setHover(compare);
    };

    // compare once and add resize listener on "componentDidMount"
    useEffect(() => {
        compareSize();
        window.addEventListener('resize', compareSize);
        return () => {
            window.removeEventListener('resize', compareSize);
        };
    }, []);

    // Define state and function to update the value
    const [hoverStatus, setHover] = useState(false);

    return (
        <Tooltip title={props?.text} interactive disableHoverListener={!hoverStatus} classes={classes} arrow>
            <div
                ref={textElementRef}
                style={{
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    width: props.width || null,
                    cursor: props.cursor || 'default',
                    textAlign: props.textAlign || null,
                    color: props.textColor || null
                }}
            >
                {props.children}
            </div>
        </Tooltip>
    );
};

export default OverflowTip;
