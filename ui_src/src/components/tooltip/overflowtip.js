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

import React, { useRef, useEffect, useState } from 'react';
import Tooltip from '@material-ui/core/Tooltip';
import { makeStyles } from '@material-ui/core/styles';

const OverflowTip = ({ color, center, width, maxWidth, cursor, textAlign, textColor, children, className, text, position }) => {
    const tooltipStyle = makeStyles((theme) => ({
        tooltip: {
            color: color === 'white' ? '#2B2E3F' : '#f7f7f7',
            cursor: 'pointer',
            backgroundColor: color === 'white' ? '#f7f7f7' : '#2B2E3F',
            fontSize: '14px',
            fontWeight: 800,
            margin: '5px',
            fontFamily: 'Inter',
            boxShadow: 'rgba(0, 0, 0, 0.24) 0px 3px 8px',
            whiteSpace: 'pre-line',
            textAlign: center ? 'center' : 'start'
        },
        arrow: {
            color: color === 'white' ? '#f7f7f7' : '#2B2E3F'
        }
    }));
    const classes = tooltipStyle();
    const textElementRef = useRef();

    const compareSize = () => {
        const compare = textElementRef.current.scrollWidth > textElementRef.current.clientWidth;
        setHover(compare);
    };

    useEffect(() => {
        compareSize();
        window.addEventListener('resize', compareSize);
        return () => {
            window.removeEventListener('resize', compareSize);
        };
    }, []);

    const [hoverStatus, setHover] = useState(false);

    return (
        <Tooltip className={className} title={text || ''} interactive disableHoverListener={!hoverStatus} classes={classes} arrow>
            <div
                ref={textElementRef}
                style={{
                    position: position,
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    width: width || null,
                    maxWidth: maxWidth || null,
                    cursor: cursor || 'default',
                    textAlign: textAlign || null,
                    color: textColor || null
                }}
            >
                {children}
            </div>
        </Tooltip>
    );
};

export default OverflowTip;
