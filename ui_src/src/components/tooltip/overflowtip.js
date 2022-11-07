// Credit for The NATS.IO Authors
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
        <Tooltip className={props?.className} title={props?.text} interactive disableHoverListener={!hoverStatus} classes={classes} arrow>
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
