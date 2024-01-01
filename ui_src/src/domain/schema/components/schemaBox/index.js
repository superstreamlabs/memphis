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

import './style.scss';

import { CloseRounded } from '@material-ui/icons';
import Drawer from "components/drawer";
import React, { useState, useEffect } from 'react';
import { useHistory } from 'react-router-dom';

import { ReactComponent as CreatedDateIcon } from 'assets/images/createdDateIcon.svg';
import { ReactComponent as NotUsedIcond } from 'assets/images/notUsedIcon.svg';
import { capitalizeFirst, parsingDate } from 'services/valueConvertor';
import CheckboxComponent from 'components/checkBox';
import { ReactComponent as UsedIcond } from 'assets/images/usedIcon.svg';
import TagsList from 'components/tagList';
import SchemaDetails from '../schemaDetails';
import OverflowTip from 'components/tooltip/overflowtip';
import pathDomains from 'router';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';

function SchemaBox({ schemaBox, handleCheckedClick, isCheck }) {
    const history = useHistory();
    const [schema, setSchema] = useState(schemaBox);
    const [open, setOpen] = useState(false);

    useEffect(() => {
        const url = window.location.href;
        const schemaName = url.split('schemaverse/list/')[1];
        if (schemaName === schema?.name) setOpen(true);
    }, []);

    useEffect(() => {
        setSchema(schemaBox);
    }, [schemaBox]);

    const handleDrawer = (flag) => {
        setOpen(flag);
        if (flag) history.push(`${pathDomains.schemaverse}/list/${schema?.name}`);
        else history.push(`${pathDomains.schemaverse}/list`);
    };

    const removeTag = async (tagName, schemaName) => {
        try {
            await httpRequest('DELETE', `${ApiEndpoints.REMOVE_TAG}`, { name: tagName, entity_type: 'schema', entity_name: schemaName });
            schema.tags = schema.tags.filter((tag) => tag.name !== tagName);
            setSchema({ ...schema });
        } catch (error) {}
    };

    const updateTags = async (tags) => {
        schema.tags = tags;
        setSchema({ ...schema });
    };

    return (
        <>
            <div>
                <CheckboxComponent checked={isCheck} id={schema.name} onChange={handleCheckedClick} name={schema.name} />
                <div key={schema.name} className="schema-box-wrapper">
                    <header is="x3d">
                        <div className="header-wrapper" onClick={() => handleDrawer(true)}>
                            <div className="schema-name">
                                <OverflowTip text={schema.name} maxWidth={'150px'}>
                                    <span>{schema.name}</span>
                                </OverflowTip>
                            </div>
                            <div className="is-used">
                                {schema.used ? (
                                    <>
                                        <UsedIcond /> <p className="used">Used</p>
                                    </>
                                ) : (
                                    <>
                                        <NotUsedIcond />
                                        <p className="not-used"> Not used</p>
                                    </>
                                )}
                            </div>
                        </div>
                    </header>
                    <type is="x3d" onClick={() => handleDrawer(true)}>
                        <div className="field-wrapper">
                            <p>Type : </p>
                            {schema.type === 'json' ? <span>JSON schema</span> : <span> {capitalizeFirst(schema.type)}</span>}
                        </div>
                        <div className="field-wrapper">
                            <p>Created by : </p>
                            <OverflowTip text={schema.created_by_username} maxWidth={'70px'}>
                                <span>{capitalizeFirst(schema.created_by_username)}</span>
                            </OverflowTip>
                        </div>
                    </type>
                    <tags is="x3d">
                        <TagsList
                            tagsToShow={3}
                            tags={schema?.tags}
                            editable
                            entityType="schema"
                            entityName={schema.name}
                            handleDelete={(tag) => removeTag(tag, schema.name)}
                            handleTagsUpdate={(tags) => updateTags(tags)}
                        />
                    </tags>
                    <date is="x3d" onClick={() => handleDrawer(true)}>
                        <CreatedDateIcon alt="createdDateIcon" />
                        <p>{parsingDate(schema.created_at)}</p>
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
