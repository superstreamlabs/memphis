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

import './style.scss';

import { CloseRounded } from '@material-ui/icons';
import { Drawer } from 'antd';
import React, { useState, useEffect } from 'react';
import { useHistory } from 'react-router-dom';

import createdDateIcon from '../../../../assets/images/createdDateIcon.svg';
import notUsedIcond from '../../../../assets/images/notUsedIcon.svg';
import { capitalizeFirst, parsingDate } from '../../../../services/valueConvertor';
import CheckboxComponent from '../../../../components/checkBox';
import usedIcond from '../../../../assets/images/usedIcon.svg';
import TagsList from '../../../../components/tagList';
import SchemaDetails from '../schemaDetails';
import OverflowTip from '../../../../components/tooltip/overflowtip';
import pathDomains from '../../../../router';

function SchemaBox({ schema, handleCheckedClick, isCheck }) {
    const history = useHistory();
    const [open, setOpen] = useState(false);

    useEffect(() => {
        const url = window.location.href;
        const schemaName = url.split('schemaverse/list/')[1];
        if (schemaName === schema?.name) setOpen(true);
    }, []);

    const handleDrawer = (flag) => {
        setOpen(flag);
        if (flag) history.push(`${pathDomains.schemaverse}/list/${schema?.name}`);
        else history.push(`${pathDomains.schemaverse}/list`);
    };

    return (
        <>
            <div>
                <CheckboxComponent checked={isCheck} id={schema.name} onChange={handleCheckedClick} name={schema.name} />
                <div key={schema.name} onClick={() => handleDrawer(true)} className="schema-box-wrapper">
                    <header is="x3d">
                        <div className="header-wrapper">
                            <div className="schema-name">
                                <OverflowTip text={schema.name} maxWidth={'150px'}>
                                    <span>{schema.name}</span>
                                </OverflowTip>
                            </div>
                            <div className="is-used">
                                <img src={schema.used ? usedIcond : notUsedIcond} alt="usedIcond" />
                                {schema.used ? <p className="used">Used</p> : <p className="not-used"> Not used</p>}
                            </div>
                        </div>
                    </header>
                    <type is="x3d">
                        <div className="field-wrapper">
                            <p>Type : </p>
                            {schema.type === 'json' ? <span>JSON schema</span> : <span> {capitalizeFirst(schema.type)}</span>}
                        </div>
                        <div className="field-wrapper">
                            <p>Created by : </p>
                            <OverflowTip text={schema.created_by_user} maxWidth={'70px'}>
                                <span>{capitalizeFirst(schema.created_by_user)}</span>
                            </OverflowTip>
                        </div>
                    </type>
                    <tags is="x3d">
                        <TagsList tagsToShow={3} tags={schema?.tags} />
                    </tags>
                    <date is="x3d">
                        <img src={createdDateIcon} alt="createdDateIcon" />
                        <p>{parsingDate(schema.creation_date)}</p>
                    </date>
                </div>
            </div>
            <Drawer
                title={schema?.name}
                placement="right"
                size={'large'}
                onClose={() => handleDrawer(false)}
                destroyOnClose={true}
                open={open}
                maskStyle={{ background: 'rgba(16, 16, 16, 0.2)' }}
                closeIcon={<CloseRounded style={{ color: '#D1D1D1' }} />}
            >
                <SchemaDetails schemaName={schema?.name} closeDrawer={() => handleDrawer(false)} />
            </Drawer>
        </>
    );
}

export default SchemaBox;
