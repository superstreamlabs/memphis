import './style.scss';

import { withStyles } from '@material-ui/core/styles';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import React, { useState } from 'react';

import { getFontColor } from '../../../utils/styleTemplates';
import comingSoonBox from '../../../assets/images/comingSoonBox.svg';
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

import ApexChart from './areaChart';

const AntTabs = withStyles({
    root: {
        position: 'absolute',
        top: '5px',
        left: '40%'
    },
    indicator: {
        backgroundColor: getFontColor('black')
    }
})(Tabs);

const AntTab = withStyles((theme) => ({
    root: {
        textTransform: 'none',
        fontSize: '14px',
        minWidth: 12,
        maxWidth: 100,
        fontWeight: theme.typography.fontWeightBold,
        marginRight: theme.spacing(3),
        fontFamily: ['Inter'].join(','),
        '&:hover': {
            color: getFontColor('navy'),
            opacity: 1
        },
        '&$selected': {
            color: getFontColor('navy'),
            fontWeight: theme.typography.fontWeightBold
        },
        '&:focus': {
            color: getFontColor('navy')
        }
    },
    selected: {}
}))((props) => <Tab disableRipple {...props} />);

const Throughput = () => {
    const [value, setValue] = useState(0);

    const handleChangeMenuItem = (_, newValue) => {
        setValue(newValue);
    };

    return (
        <div className="overview-wrapper throughput-overview-container">
            <div className="coming-soon-wrapper">
                <img src={comingSoonBox} width={40} height={70} alt="comingSoonBox" />
                <p>Coming soon</p>
            </div>
            <AntTabs value={value} onChange={handleChangeMenuItem}>
                <AntTab label="Consumers" />
                <AntTab label="Producers" />
            </AntTabs>
            <p className="overview-components-header">Throughput</p>
            <div className="throughput-chart">
                <ApexChart />
            </div>
        </div>
    );
};

export default Throughput;
