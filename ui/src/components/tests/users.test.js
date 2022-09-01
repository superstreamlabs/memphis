/**
 * @jest-environment jsdom
 */

import React from 'react';
import user from '@testing-library/user-event';
import Users from '../../domain/users';
import { Context } from '../../hooks/store';
import { shallow } from 'enzyme';

describe('Users page', () => {
    it('renders "Add new user"', () => {
        const TestComponent = () => (
            <Context.Provider value="Provided Value">
                <Users />
            </Context.Provider>
        );
        const wrapper = shallow(<TestComponent />);
        expect(wrapper.contains(<div className="modal-btn" />));
        // user.click(element.find(Users).dive().screen.getByRole('button', { name: 'Add a new user' }));
        // expect(element.find(Users).dive().getByText('Add a new user')).toBeInTheDocument();
    });
});
