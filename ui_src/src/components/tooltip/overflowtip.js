// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

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
                    maxWidth: props.maxWidth || null,
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
