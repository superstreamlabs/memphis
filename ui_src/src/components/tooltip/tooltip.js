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
