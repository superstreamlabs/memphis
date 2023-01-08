import './style.scss';

import { withStyles } from '@material-ui/core/styles';
import Tabs from '@material-ui/core/Tabs';
import Tab from '@material-ui/core/Tab';
import React, { useState } from 'react';

import { getFontColor } from '../../../utils/styleTemplates';
import comingSoonBox from '../../../assets/images/comingSoonBox.svg';

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
